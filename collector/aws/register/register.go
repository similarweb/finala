package register

import (
	"finala/collector/aws/common"

	log "github.com/sirupsen/logrus"
)

// resourcesList includes all registered resources
var resourcesList = map[string]common.DetectResourceMaker{}

// Registry add new resource to execute
func Registry(name string, resourceInit common.DetectResourceMaker) {
	log.WithField("resource", name).Debug("Registry resource")
	resourcesList[name] = resourceInit
}

// GetResources returns all registered resources
func GetResources() map[string]common.DetectResourceMaker {
	return resourcesList
}
