package aws

import (
	"encoding/json"
	"finala/collector"
	"fmt"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// EC2VolumeClientDescriptor is an interface defining the AWS EC2
type EC2VolumeClientDescriptor interface {
	DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
}

// EC2VolumeManager describe EBS manager
type EC2VolumeManager struct {
	collector          collector.CollectorDescriber
	client             EC2VolumeClientDescriptor
	pricingClient      *PricingManager
	region             string
	namespace          string
	servicePricingCode string
	Type               string
}

// DetectedAWSEC2Volume define the detected volume data
type DetectedAWSEC2Volume struct {
	Region        string
	ResourceID    string
	Type          string
	Size          int64
	PricePerMonth float64 `gorm:"type:DOUBLE`
	Tags          string  `gorm:"type:TEXT" json:"-"`
}

// NewVolumesManager implements AWS GO SDK
func NewVolumesManager(collector collector.CollectorDescriber, client EC2VolumeClientDescriptor, pricing *PricingManager, region string) *EC2VolumeManager {

	return &EC2VolumeManager{
		collector:          collector,
		client:             client,
		pricingClient:      pricing,
		region:             region,
		namespace:          "AWS/EC2",
		servicePricingCode: "AmazonEC2",
		Type:               fmt.Sprintf("%s_ec2_volume", ResourcePrefix),
	}
}

// Detect unused volumes
func (ev *EC2VolumeManager) Detect() ([]DetectedAWSEC2Volume, error) {

	log.Info("analyze Volumes")

	detected := []DetectedAWSEC2Volume{}
	volumes, err := ev.Describe(nil, nil)

	if err != nil {
		log.WithField("error", err).Error("could not describe ec2 volumes")
		return detected, err
	}

	// Set storage filters to for pricing API
	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("productFamily"),
			Value: awsClient.String("Storage"),
		},
	}

	for _, vol := range volumes {

		log.WithField("id", *vol.VolumeId).Info("Volume found")

		price, err := ev.pricingClient.GetPrice(ev.GetBasePricingFilterInput(vol, filters), "")
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume_id": *vol.VolumeId,
				"filters":   filters,
			}).Error("Error when trying to get volume price")
			price = 0
		}

		decodedTags := []byte{}
		if err == nil {
			decodedTags, err = json.Marshal(vol.Tags)
		}

		volumeSize := *vol.Size
		dEBS := DetectedAWSEC2Volume{
			Region:        ev.region,
			ResourceID:    *vol.VolumeId,
			Type:          *vol.VolumeType,
			Size:          volumeSize,
			PricePerMonth: ev.GetCalculatedPrice(vol, price),
			Tags:          string(decodedTags),
		}

		ev.collector.Add(collector.EventCollector{
			Name: "resource-detected",
			Data: collector.ResourceDetected{
				ResourceName: ev.Type,
				Data:         dEBS,
			},
		})

		detected = append(detected, dEBS)

	}

	return detected, nil

}

// GetCalculatedPrice calculate the volume price by volume type
func (ev *EC2VolumeManager) GetCalculatedPrice(vol *ec2.Volume, basePrice float64) float64 {

	volumeSize := *vol.Size
	switch *vol.VolumeType {
	case "io1":
		// For io1 we need to add IOPs to the price (https://aws.amazon.com/ebs/pricing).
		extraFilter := []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String("EBS:VolumeP-IOPS.piops"),
			},
		}

		iopsPrice, err := ev.pricingClient.GetPrice(ev.GetBasePricingFilterInput(vol, extraFilter), "")
		if err != nil {
			iopsPrice = 0
		}
		return basePrice*float64(volumeSize) + iopsPrice*float64(*vol.Iops)
	default:
		return basePrice * float64(volumeSize)
	}

}

// GetBasePricingFilterInput set the pricing product filters
func (ev *EC2VolumeManager) GetBasePricingFilterInput(vol *ec2.Volume, extraFilters []*pricing.Filter) *pricing.GetProductsInput {

	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("volumeApiName"),
			Value: awsClient.String(*vol.VolumeType),
		},
	}

	if extraFilters != nil {
		filters = append(filters, extraFilters...)
	}

	return &pricing.GetProductsInput{
		ServiceCode: &ev.servicePricingCode,
		Filters:     filters,
	}

}

// Describe return list of volumes with available status
func (ev *EC2VolumeManager) Describe(token *string, volumes []*ec2.Volume) ([]*ec2.Volume, error) {

	input := &ec2.DescribeVolumesInput{
		NextToken: token,
		Filters: []*ec2.Filter{
			{
				Name:   awsClient.String("status"),
				Values: []*string{awsClient.String("available"), awsClient.String("error")},
			},
		},
	}

	resp, err := ev.client.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}

	if volumes == nil {
		volumes = []*ec2.Volume{}
	}

	for _, vol := range resp.Volumes {
		volumes = append(volumes, vol)
	}
	if resp.NextToken != nil {
		ev.Describe(resp.NextToken, volumes)
	}

	return volumes, nil
}
