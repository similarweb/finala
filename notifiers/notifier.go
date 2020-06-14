package notifiers

import (
	"encoding/json"
	notifierCommon "finala/notifiers/common"
	"net/url"

	"finala/request"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type NotifierMaker func() notifierCommon.Notifier

const notRegisteredTemplate = "notifier by the name %s was not registered"

var registeredNotifiers = map[notifierCommon.NotifierName]NotifierMaker{}

// DataFetcherManager will hold all the data for Finala notifier
type DataFetcherManager struct {
	client      request.HTTPClientDescriber
	log         log.Entry
	apiEndpoint string
}

// NewDataFetcherManager will fetch all the data requests from Finala API.
func NewDataFetcherManager(client request.HTTPClientDescriber, log log.Entry, apiEndpoint string) *DataFetcherManager {
	return &DataFetcherManager{
		client:      client,
		log:         log,
		apiEndpoint: apiEndpoint,
	}
}

// GetLatestExecution will get the Collector's latest execution
func (dfm *DataFetcherManager) GetLatestExecution() (latestExecution string, err error) {
	req, err := dfm.client.Request("GET", fmt.Sprintf("%s/api/v1/executions?querylimit=1", dfm.apiEndpoint), nil, nil)
	if err != nil {
		dfm.log.WithError(err).Error("could not create HTTP client request")
		return "", err
	}

	res, err := dfm.client.DO(req)
	if err != nil {
		dfm.log.WithError(err).Error("could not send HTTP client request")
		return "", err
	}

	defer res.Body.Close()

	var executions []notifierCommon.NotifierExecutionsResponse
	err = json.NewDecoder(res.Body).Decode(&executions)

	return executions[0].ID, err
}

// GetExecutionSummary will get the Collector's execution summary by given filters
func (dfm *DataFetcherManager) GetExecutionSummary(executionID string, filterOptions map[string]string) (map[string]*notifierCommon.NotifierCollectorsSummary, error) {
	v := url.Values{}
	for filterName, filterValue := range filterOptions {
		v.Set(filterName, filterValue)
	}
	req, err := dfm.client.Request("GET", fmt.Sprintf("%s/api/v1/summary/%s", dfm.apiEndpoint, executionID), v, nil)
	if err != nil {
		dfm.log.WithError(err).Error("could not create HTTP client request")
		return nil, err
	}

	res, err := dfm.client.DO(req)
	if err != nil {
		dfm.log.WithError(err).Error("could not send HTTP client request")
		return nil, err
	}

	defer res.Body.Close()

	var executionSummary map[string]*notifierCommon.NotifierCollectorsSummary
	err = json.NewDecoder(res.Body).Decode(&executionSummary)

	return executionSummary, err
}
