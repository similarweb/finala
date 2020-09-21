package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"finala/api/config"
	"finala/api/storage"
	"finala/interpolation"
	"fmt"
	"reflect"
	"strconv"
	"time"

	elastic "github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

var (
	ErrInvalidQuery            = errors.New("invalid query")
	ErrAggregationTermNotFound = errors.New("aggregation terms was not found")
)

const (

	// prefixDayIndex defins the index name of the current day
	prefixIndexName = "finala-%s"

	// indexMapping define the default index mapping
	indexMapping = `{
		"mappings":{
			"properties":{
				"ResourceName":{
					"type":"keyword"
				},
				"ExecutionID":{
					"type":"keyword"
				},
				"EventType":{
					"type":"keyword"
				},
				"Timestamp":{
					"type":"date"
				}
			}
		}
	}`
)

// StorageManager descrive elasticsearchStorage
type StorageManager struct {
	client          elasticSearchDescriptor
	currentIndexDay string
}

// NewStorageManager creates new elasticsearch storage
func NewStorageManager(conf config.ElasticsearchConfig) (*StorageManager, error) {

	client, err := NewClient(conf)
	if err != nil {
		return nil, err
	}

	storageManager := &StorageManager{
		client: client,
	}

	if !storageManager.setCreateCurrentIndexDay() {
		return nil, errors.New("could not create index")
	}
	go func() {
		for {
			now := time.Now().In(time.UTC)
			diff := storageManager.getDurationUntilTomorrow(now)
			log.WithFields(log.Fields{
				"now":      now,
				"duration": diff,
			}).Info("change index in")
			// wait until duration end
			<-time.After(diff)
			storageManager.setCreateCurrentIndexDay()
		}
	}()

	return storageManager, nil
}

// Save new documents
func (sm *StorageManager) Save(data string) bool {

	_, err := sm.client.Index().
		Index(sm.currentIndexDay).
		BodyJson(data).
		Do(context.Background())

	if err != nil {
		log.WithFields(log.Fields{
			"index": sm.currentIndexDay,
			"data":  data,
		}).WithError(err).Error("Fail to save document")
		return false
	}

	return true

}

// getDynamicMatchQuery will iterate through a filters map and create Match Query for each of them
func (sm *StorageManager) getDynamicMatchQuery(filters map[string]string, operator string) []elastic.Query {
	dynamicMatchQuery := []elastic.Query{}
	var mq *elastic.MatchQuery
	for name, value := range filters {
		mq = elastic.NewMatchQuery(name, value)
		// Minimum number of clauses that must match for a document to be returned
		mq.MinimumShouldMatch("100%")
		if operator == "and" {
			mq = mq.Operator("and")
		}

		dynamicMatchQuery = append(dynamicMatchQuery, mq)
	}
	return dynamicMatchQuery
}

// GetSummary returns executions summary
func (sm *StorageManager) GetSummary(executionID string, filters map[string]string) (map[string]storage.CollectorsSummary, error) {
	summary := map[string]storage.CollectorsSummary{}
	executionIDQuery := elastic.NewTermQuery("ExecutionID", executionID)
	eventTypeQuery := elastic.NewTermQuery("EventType", "service_status")

	log.WithFields(log.Fields{
		"execution_id": executionIDQuery,
		"event_type":   eventTypeQuery,
	}).Debug("Going to get get summary with the following fields")

	searchQuery := sm.client.Search().
		Query(elastic.NewBoolQuery().Must(eventTypeQuery, executionIDQuery))

	searchResultTotalHits, err := searchQuery.Size(0).Do(context.Background())

	if err != nil {
		log.WithError(err).Error("error when trying get total hits summary")
		return summary, err
	}

	searchResult, err := searchQuery.Size(int(searchResultTotalHits.TotalHits())).Do(context.Background())

	if err != nil {
		log.WithError(err).Error("error when trying to get summary data")
		return summary, err
	}

	log.WithFields(log.Fields{
		"milliseconds": searchResult.TookInMillis,
		"hits":         len(searchResult.Hits.Hits),
	}).Debug("get all executions id response time")

	var summaryData storage.Summary
	for _, item := range searchResult.Each(reflect.TypeOf(summaryData)) {
		summaryRow, ok := item.(storage.Summary)
		if !ok {
			log.Error("could not parse summary row")
			continue
		}

		// check if the resource status already exists, if yes we check if have latest event
		val, found := summary[summaryRow.ResourceName]
		if found {
			if summaryRow.EventTime < val.EventTime {
				continue
			}
			delete(summary, summaryRow.ResourceName)
		}

		summary[summaryRow.ResourceName] = storage.CollectorsSummary{
			EventTime:    summaryRow.EventTime,
			Status:       summaryRow.Data.Status,
			ResourceName: summaryRow.ResourceName,
			ErrorMessage: summaryRow.Data.ErrorMessage,
		}
	}

	for resourceName, resourceData := range summary {
		filters["ResourceName"] = resourceName
		log.WithField("filters", filters).Debug("Going to get resources summary details with the following filters")
		totalSpent, resourceCount, err := sm.getResourceSummaryDetails(executionID, filters)

		if err != nil {
			continue
		}
		newResourceData := resourceData
		newResourceData.TotalSpent = totalSpent
		newResourceData.ResourceCount = resourceCount
		summary[resourceName] = newResourceData

	}

	return summary, nil

}

// getResourceSummaryDetails returns total resource spent and total resources detected
func (sm *StorageManager) getResourceSummaryDetails(executionID string, filters map[string]string) (float64, int64, error) {

	var totalSpent float64
	var resourceCount int64

	dynamicMatchQuery := sm.getDynamicMatchQuery(filters, "or")
	dynamicMatchQuery = append(dynamicMatchQuery, elastic.NewTermQuery("ExecutionID", executionID))
	dynamicMatchQuery = append(dynamicMatchQuery, elastic.NewTermQuery("EventType", "resource_detected"))

	searchResult, err := sm.client.Search().
		Query(elastic.NewBoolQuery().Must(dynamicMatchQuery...)).
		Aggregation("sum", elastic.NewSumAggregation().Field("Data.PricePerMonth")).
		Size(0).Do(context.Background())

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"filters": filters,
		}).Error("error when trying to get summary details")

		return totalSpent, resourceCount, err
	}

	log.WithFields(log.Fields{
		"filters":      filters,
		"milliseconds": searchResult.TookInMillis,
	}).Debug("get execution details")

	resp, ok := searchResult.Aggregations.Terms("sum")
	if ok {
		if val, ok := resp.Aggregations["value"]; ok {

			totalSpent, _ = strconv.ParseFloat(string(val), 64)
			resourceCount = searchResult.Hits.TotalHits.Value
		}
	}

	return totalSpent, resourceCount, nil
}

// GetExecutions returns collector executions
func (sm *StorageManager) GetExecutions(queryLimit int) ([]storage.Executions, error) {
	executions := []storage.Executions{}

	// First search for all message with eventType: service_status
	// Second look for message which have the field ExecutionID
	// Third Order the ExecutionID by EventTime Desc
	searchResult, err := sm.client.Search().Aggregation("orderedExecutionID", elastic.NewFiltersAggregation().
		Filters(elastic.NewBoolQuery().Filter(elastic.NewBoolQuery().Should(elastic.NewTermQuery("EventType", "service_status")))).
		SubAggregation("ExecutionIDDesc", elastic.NewTermsAggregation().Field("ExecutionID").Size(queryLimit).Order("MaxEventTime", false).
			SubAggregation("MaxEventTime", elastic.NewMaxAggregation().Field("EventTime")))).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("error when trying to get executions collectors")
		return executions, ErrInvalidQuery
	}

	resp, ok := searchResult.Aggregations.Terms("orderedExecutionID")
	if !ok {
		log.Error("orderedExecutionID field term does not exist")
		return executions, ErrAggregationTermNotFound
	}

	for _, ExecutionIDBuckets := range resp.Buckets {
		descOrderedExecutionIDs := ExecutionIDBuckets.Aggregations["ExecutionIDDesc"]

		var executionsIDs orderedExecutionIDs
		err := json.Unmarshal([]byte(string(descOrderedExecutionIDs)), &executionsIDs)
		if err != nil {
			log.WithError(err).Error("error when trying to parse bucket aggregations execution ids")
			return executions, nil
		}

		for _, executionIDValue := range executionsIDs.Buckets {
			executionID := string(executionIDValue.Key)

			executionName, err := interpolation.ExtractExecutionName(executionID)
			if err != nil {
				log.WithError(err).WithField("execution_name", executionName).Error("could not extract execution name")
				continue
			}

			collectorExecutionTime, err := interpolation.ExtractTimestamp(executionID)
			if err != nil {
				log.WithError(err).WithField("collector_execution_time", collectorExecutionTime).Error("could not parse to int64")
				continue
			}

			executions = append(executions, storage.Executions{
				ID:   executionID,
				Name: executionName,
				Time: time.Unix(collectorExecutionTime, 0),
			})
		}
	}
	return executions, nil
}

// GetResources return resource data
func (sm *StorageManager) GetResources(resourceType string, executionID string, filters map[string]string) ([]map[string]interface{}, error) {

	var resources []map[string]interface{}
	dynamicMatchQuery := sm.getDynamicMatchQuery(filters, "or")
	componentQ := elastic.NewTermQuery("EventType", "resource_detected")
	deploymentQ := elastic.NewTermQuery("ExecutionID", executionID)
	ResourceNameQ := elastic.NewTermQuery("ResourceName", resourceType)
	generalQ := elastic.NewBoolQuery()
	generalQ = generalQ.Must(componentQ).Must(deploymentQ).Must(ResourceNameQ).Must(dynamicMatchQuery...)
	searchResultTotalHits, err := sm.client.Search().
		Query(generalQ).
		Pretty(true).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("elasticsearch query error")
		return resources, err
	}

	searchResult, err := sm.client.Search().
		Query(generalQ).
		Pretty(true).
		Size(int(searchResultTotalHits.TotalHits())).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("elasticsearch query error")
		return resources, err
	}

	for _, hit := range searchResult.Hits.Hits {

		rowData := make(map[string]interface{})
		err := json.Unmarshal([]byte(string(hit.Source)), &rowData)
		if err != nil {
			log.WithError(err).Error("error when trying to parse search result hits data")
			continue
		}

		resources = append(resources, rowData)
	}

	return resources, nil
}

// GetResourceTrends return resource data
func (sm *StorageManager) GetResourceTrends(resourceType string, filters map[string]string, limit int) ([]storage.ExecutionCost, error) {
	// Per resource trends, filters should take care of granularity (per resource: Data.ResourceID, Data.Region, Data.Metric -> Data.PricePerMonth)
	var resources []storage.ExecutionCost
	var mustNotQuery []elastic.Query

	// Must
	mustQuery := sm.getDynamicMatchQuery(filters, "and")
	mustQuery = append(mustQuery, elastic.NewTermQuery("ResourceName", resourceType))

	// Unsupported Types - MustNot
	mustNotQuery = append(mustNotQuery, elastic.NewTermQuery("EventType", "service_status"))
	mustNotQuery = append(mustNotQuery, elastic.NewTermQuery("ResourceName", "aws_iam_users"))
	mustNotQuery = append(mustNotQuery, elastic.NewTermQuery("ResourceName", "aws_elastic_ip"))
	mustNotQuery = append(mustNotQuery, elastic.NewTermQuery("ResourceName", "aws_lambda"))
	mustNotQuery = append(mustNotQuery, elastic.NewTermQuery("ResourceName", "aws_ec2_volume"))

	queryBuilder := elastic.NewBoolQuery().MustNot(mustNotQuery...).Must(mustQuery...)
	searchResult, err := sm.client.Search().
		Query(queryBuilder).
		Pretty(true).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("elasticsearch query size error")
		return resources, err
	}

	searchResultQuerySize := int(searchResult.TotalHits())
	searchResult, err = sm.client.Search().
		Query(queryBuilder).
		Pretty(true).
		Size(searchResultQuerySize).
		SortBy(elastic.NewFieldSort("Timestamp").Desc()).
		Aggregation("executions", elastic.NewTermsAggregation().Field("ExecutionID").OrderByKeyDesc(). // Aggregate by ExecutionID
														SubAggregation("monthly-cost", elastic.NewSumAggregation().Field("Data.PricePerMonth"))). // Sub aggregate and sum by Data.PricePerMonth per bucket
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("elasticsearch query error")
		return resources, err
	}

	executions, found := searchResult.Aggregations.Terms("executions")
	if found {
		for _, ppm := range executions.Buckets {
			executionID := ppm.Key.(string)
			monthlyAgg, _ := ppm.Aggregations.Sum("monthly-cost")

			// Extract the timestamp from the ExecutionID
			timestamp, err := interpolation.ExtractTimestamp(executionID)
			if err != nil {
				timestamp = 0
			}

			resources = append(resources, storage.ExecutionCost{
				ExecutionID:        executionID,
				ExtractedTimestamp: timestamp,
				CostSum:            *monthlyAgg.Value,
			})
		}
	}

	// Maximum number of resources to return
	if len(resources) > limit {
		resources = resources[0:limit]
	}

	return resources, nil
}

// GetExecutionTags will return the tags according to a given executionID
func (sm *StorageManager) GetExecutionTags(executionID string) (map[string][]string, error) {

	tags := map[string][]string{}
	eventTypeMatchQuery := elastic.NewTermQuery("EventType", "resource_detected")
	executionIDMatchQuery := elastic.NewTermQuery("ExecutionID", executionID)
	elasticQuery := elastic.Query(elastic.NewBoolQuery().Must(eventTypeMatchQuery, executionIDMatchQuery))

	// First get the Query Size for all the hits
	searchResultHits, err := sm.client.Search().
		Query(elasticQuery).
		Pretty(true).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("got an elasticsearch error while running the query to get the hits number")
		return tags, err
	}

	searchResultQuerySize := int(searchResultHits.TotalHits())

	// Second query with the size of the hits
	searchResult, err := sm.client.Search().
		Query(elasticQuery).
		Pretty(true).
		Size(searchResultQuerySize).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("got an elasticsearch error while running the query")
		return tags, err
	}

	var availableTags TagsData
	for _, hit := range searchResult.Hits.Hits {

		err := json.Unmarshal([]byte(string(hit.Source)), &availableTags)
		if err != nil {
			log.WithError(err).Error("error when trying to parse tags map")
			continue
		}

		for key, value := range availableTags.Data.Tag {
			tags[key] = append(tags[key], value)
		}
	}

	// Make sure the values of each tag unique
	for tagName, tagValues := range tags {
		tags[tagName] = interpolation.UniqueStr(tagValues)
	}

	return tags, nil
}

// createIndex creating create elasticsearch index if not exists
func (sm *StorageManager) createIndex(index string) error {

	exists, err := sm.client.IndexExists(index).Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"index": index,
		}).WithError(err).Error("Error when trying to check if elasticsearch exists")
		return err
	}

	if exists {
		log.WithField("index", index).Info("index already exists")
		return nil
	}

	ctx := context.Background()
	_, err = sm.client.CreateIndex(index).BodyString(indexMapping).Do(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"index": index,
		}).WithError(err).Error("Error when trying to create elasticsearch index")
		return err
	}

	log.WithField("index", index).Info("index created successfully")
	return nil

}

// getDurationUntilTomorrow returns the duration time until tomorrow
func (sm *StorageManager) getDurationUntilTomorrow(now time.Time) time.Duration {

	zone, _ := now.Zone()
	location, err := time.LoadLocation(zone)
	if err != nil {
		log.WithError(err).WithField("zone", zone).Warn("zone name not found")
		location = time.UTC
	}

	tomorrow := getDayAfterDate(now, location)
	diff := tomorrow.Sub(now)

	return diff

}

// setCreateCurrentIndexDay create and set the current day as index
func (sm *StorageManager) setCreateCurrentIndexDay() bool {
	dt := time.Now().In(time.UTC)
	newIndex := fmt.Sprintf(prefixIndexName, dt.Format("01-02-2006"))
	log.WithFields(log.Fields{
		"current_index_day":    sm.currentIndexDay,
		"to_current_index_day": newIndex,
	}).Info("change current index day")
	err := sm.createIndex(newIndex)
	if err != nil {
		return false
	}

	sm.currentIndexDay = newIndex
	return true

}
