package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"

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
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedAWSEC2Volume define the detected volume data
type DetectedAWSEC2Volume struct {
	Metric        string
	Region        string
	ResourceID    string
	Type          string
	Size          int64
	PricePerMonth float64
	Tag           map[string]string
}

func init() {
	register.Registry("ec2_volumes", NewVolumesManager)
}

// NewVolumesManager implements AWS GO SDK
func NewVolumesManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = ec2.New(awsManager.GetSession())
	}

	ec2Client, ok := client.(EC2VolumeClientDescriptor)
	if !ok {
		return nil, errors.New("invalid ec2 volumes client")
	}

	return &EC2VolumeManager{
		client:             ec2Client,
		awsManager:         awsManager,
		namespace:          "AWS/EC2",
		servicePricingCode: "AmazonEC2",
		Name:               awsManager.GetResourceIdentifier("ec2_volumes"),
	}, nil

}

// Detect unused volumes
func (ev *EC2VolumeManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	// This resource support only one metric
	metric := metrics[0]

	log.WithFields(log.Fields{
		"region":   ev.awsManager.GetRegion(),
		"resource": "ec2_volume",
	}).Info("starting to analyze resource")

	ev.awsManager.GetCollector().CollectStart(ev.Name)

	detected := []DetectedAWSEC2Volume{}
	volumes, err := ev.describe(nil, nil)

	if err != nil {
		log.WithField("error", err).Error("could not describe ec2 volumes")
		ev.awsManager.GetCollector().CollectError(ev.Name, err)
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

		price, err := ev.awsManager.GetPricingClient().GetPrice(ev.getBasePricingFilterInput(vol, filters), "", ev.awsManager.GetRegion())
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
			Region:        ev.awsManager.GetRegion(),
			Metric:        metric.Description,
			ResourceID:    *vol.VolumeId,
			Type:          *vol.VolumeType,
			Size:          volumeSize,
			PricePerMonth: ev.getCalculatedPrice(vol, price),
			Tag:           tagsData,
		}

		ev.awsManager.GetCollector().AddResource(collector.EventCollector{
			ResourceName: ev.Name,
			Data:         dEBS,
		})

		detected = append(detected, dEBS)

	}

	ev.awsManager.GetCollector().CollectFinish(ev.Name)

	return detected, nil

}

// getCalculatedPrice calculate the volume price by volume type
func (ev *EC2VolumeManager) getCalculatedPrice(vol *ec2.Volume, basePrice float64) float64 {

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

		iopsPrice, err := ev.awsManager.GetPricingClient().GetPrice(ev.getBasePricingFilterInput(vol, extraFilter), "", ev.awsManager.GetRegion())
		if err != nil {
			iopsPrice = 0
		}
		return basePrice*float64(volumeSize) + iopsPrice*float64(*vol.Iops)
	default:
		return basePrice * float64(volumeSize)
	}

}

// getBasePricingFilterInput set the pricing product filters
func (ev *EC2VolumeManager) getBasePricingFilterInput(vol *ec2.Volume, extraFilters []*pricing.Filter) pricing.GetProductsInput {

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

	return pricing.GetProductsInput{
		ServiceCode: &ev.servicePricingCode,
		Filters:     filters,
	}

}

// describe return list of volumes with available status
func (ev *EC2VolumeManager) describe(token *string, volumes []*ec2.Volume) ([]*ec2.Volume, error) {

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
		return ev.describe(resp.NextToken, volumes)
	}

	return volumes, nil
}
