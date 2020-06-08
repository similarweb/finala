package notifiers

import (
	"encoding/json"
	"finala/api/storage"
	notifierCommon "finala/notifiers/common"
	"net/url"

	"finala/request"
	"fmt"

	"github.com/apex/log"
)

type NotifierMaker func() notifierCommon.Notifier

const notRegisteredTemplate = "notifier by the name %s was not registered"

var registeredNotifiers = map[notifierCommon.NotifierName]NotifierMaker{}

// NotifierManager will hold all the data for Finala notifier
type NotifierManager struct {
	RegisteredNotifiers []notifierCommon.Notifier
	client              request.HTTPClientDescriber
	apiEndpoint         string
}

// NewNotifierManager will manage all registered notifiers and requests to Finala API.
func NewNotifierManager(registeredNotifiers []notifierCommon.Notifier, client request.HTTPClientDescriber, apiEndpoint string) *NotifierManager {
	return &NotifierManager{
		RegisteredNotifiers: registeredNotifiers,
		client:              client,
		apiEndpoint:         apiEndpoint,
	}
}

// GetLatestExecution will get the Collector's latest execution
func (nfm *NotifierManager) GetLatestExecution() ([]*storage.Executions, error) {
	req, err := nfm.client.Request("GET", fmt.Sprintf("%s/api/v1/executions?querylimit=1", nfm.apiEndpoint), nil, nil)
	if err != nil {
		log.WithError(err).Error("could not create HTTP client request")
		return nil, err
	}

	res, err := nfm.client.DO(req)
	if err != nil {
		log.WithError(err).Error("could not send HTTP client request")
		return nil, err
	}

	defer res.Body.Close()

	var executions []*storage.Executions
	err = json.NewDecoder(res.Body).Decode(&executions)

	return executions, nil
}

// GetExecutionSummary will get the Collector's execution summary by given filters
func (nfm *NotifierManager) GetExecutionSummary(filterOptions map[string]string) (map[string]*storage.CollectorsSummary, error) {
	v := url.Values{}
	for filterName, filterValue := range filterOptions {
		v.Set(filterName, filterValue)
	}
	req, err := nfm.client.Request("GET", fmt.Sprintf("%s/api/v1/summary", nfm.apiEndpoint), v, nil)
	if err != nil {
		log.WithError(err).Error("could not create HTTP client request")
		return nil, err
	}

	res, err := nfm.client.DO(req)
	if err != nil {
		log.WithError(err).Error("could not send HTTP client request")
		return nil, err
	}

	defer res.Body.Close()

	var executionSummary map[string]*storage.CollectorsSummary
	err = json.NewDecoder(res.Body).Decode(&executionSummary)

	return executionSummary, nil
}
