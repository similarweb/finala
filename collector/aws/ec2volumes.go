package aws

import (
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
	client             EC2VolumeClientDescriptor
	pricingClient      *PricingManager
	collector          collector.CollectorDescriber
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedAWSEC2Volume define the detected volume data
type DetectedAWSEC2Volume struct {
	Region        string
	ResourceID    string
	Type          string
	Size          int64
	PricePerMonth float64
	Tag           map[string]string
}

// NewVolumesManager implements AWS GO SDK
func NewVolumesManager(collector collector.CollectorDescriber, client EC2VolumeClientDescriptor, pricing *PricingManager, region string) *EC2VolumeManager {

	return &EC2VolumeManager{
		client:        client,
		pricingClient: pricing,
		region:        region,
		collector:     collector,

		namespace:          "AWS/EC2",
		servicePricingCode: "AmazonEC2",
		Name:               fmt.Sprintf("%s_ec2_volume", ResourcePrefix),
	}
}

// Detect unused volumes
func (ev *EC2VolumeManager) Detect() ([]DetectedAWSEC2Volume, error) {

	log.WithFields(log.Fields{
		"region":   ev.region,
		"resource": "ec2_volume",
	}).Info("starting to analyze resource")

	ev.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ev.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detected := []DetectedAWSEC2Volume{}
	volumes, err := ev.Describe(nil, nil)

	if err != nil {
		log.WithField("error", err).Error("could not describe ec2 volumes")
		ev.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: ev.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})
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

		log.WithField("id", *vol.VolumeId).Debug("cheking ec2 volume")

		price, err := ev.pricingClient.GetPrice(ev.GetBasePricingFilterInput(vol, filters), "", ev.region)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"volume_id": *vol.VolumeId,
				"filters":   filters,
			}).Error("Error when trying to get volume price")
			price = 0
		}

		tagsData := map[string]string{}
		if err == nil {
			for _, tag := range vol.Tags {
				tagsData[*tag.Key] = *tag.Value
			}
		}

		volumeSize := *vol.Size
		dEBS := DetectedAWSEC2Volume{
			Region:        ev.region,
			ResourceID:    *vol.VolumeId,
			Type:          *vol.VolumeType,
			Size:          volumeSize,
			PricePerMonth: ev.GetCalculatedPrice(vol, price),
			Tag:           tagsData,
		}

		ev.collector.AddResource(collector.EventCollector{
			ResourceName: ev.Name,
			Data:         dEBS,
		})

		detected = append(detected, dEBS)

	}

	ev.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ev.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

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

		iopsPrice, err := ev.pricingClient.GetPrice(ev.GetBasePricingFilterInput(vol, extraFilter), "", ev.region)
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

	volumes = append(volumes, resp.Volumes...)

	if resp.NextToken != nil {
		return ev.Describe(resp.NextToken, volumes)
	}

	return volumes, nil
}
