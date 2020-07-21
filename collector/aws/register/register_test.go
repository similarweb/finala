package register

import (
	"finala/collector/aws/common"
	"finala/collector/config"
	"testing"
)

type mockResource struct {
}

func newMockResource(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	return &mockResource{}, nil
}

func (mr *mockResource) Detect(metrics []config.MetricConfig) (interface{}, error) {

	data := []string{"foo"}
	return data, nil

}

func TestRegister(t *testing.T) {

	Registry("foo", newMockResource)

	resources := GetResources()

	if len(resources) != 1 {
		t.Fatalf("unexpected resource  count, got %d expected %d", len(resources), 1)

	}

	_, exists := resources["foo"]
	if !exists {
		t.Fatalf("unexpected resources data foo doesn't exist")

	}

}
