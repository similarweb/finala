package aws

import (
	"encoding/json"
	"finala/storage"

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
	storage            storage.Storage
	pricingClient      *PricingManager
	region             string
	namespace          string
	servicePricingCode string
}

// DetectedAWSEC2Volume define the detected volume data
type DetectedAWSEC2Volume struct {
	Region        string
	ID            string
	Type          string
	Size          int64
	PricePerMonth float64 `gorm:"type:DOUBLE`
	Tags          string  `gorm:"type:TEXT" json:"-"`
}

// TableName will set the table name to storage interface
func (DetectedAWSEC2Volume) TableName() string {
	return "aws_ec2_volume"
}

// NewVolumesManager implements AWS GO SDK
func NewVolumesManager(client EC2VolumeClientDescriptor, st storage.Storage, pricing *PricingManager, region string) *EC2VolumeManager {

	st.AutoMigrate(&DetectedAWSEC2Volume{})

	return &EC2VolumeManager{
		client:        client,
		storage:       st,
		pricingClient: pricing,
		region:        region,

		namespace:          "AWS/EC2",
		servicePricingCode: "AmazonEC2",
	}
}

// Detect unused volumes
func (ev *EC2VolumeManager) Detect() ([]DetectedAWSEC2Volume, error) {

	log.Info("analyze Volumes")

	detected := []DetectedAWSEC2Volume{}
	volumes, err := ev.Describe()

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
			ID:            *vol.VolumeId,
			Type:          *vol.VolumeType,
			Size:          volumeSize,
			PricePerMonth: ev.GetCalculatedPrice(vol, price),
			Tags:          string(decodedTags),
		}

		detected = append(detected, dEBS)
		ev.storage.Create(&dEBS)

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
func (ev *EC2VolumeManager) Describe() ([]*ec2.Volume, error) {

	input := &ec2.DescribeVolumesInput{
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

	volumes := []*ec2.Volume{}
	for _, vol := range resp.Volumes {
		volumes = append(volumes, vol)

	}

	return volumes, nil
}
