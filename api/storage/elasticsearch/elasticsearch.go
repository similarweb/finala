package elasticsearch

import (
	"context"
	"encoding/json"
	"finala/api/config"
	"finala/api/storage"
	"finala/interpolation"
	"reflect"
	"strconv"
	"strings"
	"time"

	elastic "github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

const (
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
	client       *elastic.Client
	defaultIndex string
}

// getESClient create new elasticsearch client
func getESClient(conf config.ElasticsearchConfig) (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL(strings.Join(conf.Endpoints, ",")),
		elastic.SetErrorLog(log.New()),
		//elastic.SetTraceLog(log.New()), // Uncomment for debugging ElasticSearch Queries
		elastic.SetBasicAuth(conf.Username, conf.Password),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(true))

	return client, err

}

// NewStorageManager creates new elasticsearch storage
func NewStorageManager(conf config.ElasticsearchConfig) (*StorageManager, error) {

	var esclient *elastic.Client

	c := make(chan int, 1)
	go func() {
		var err error
		for {
			esclient, err = getESClient(conf)
			if err == nil {
				break
			}
			log.WithFields(log.Fields{
				"endpoint": conf.Endpoints,
			}).WithError(err).Warn("could not initialize connection to elasticsearch, retrying in 5 seconds")
			time.Sleep(5 * time.Second)
		}
		c <- 1
	}()

	select {
	case <-c:
	case <-time.After(60 * time.Second):
		log.Fatal("could not connect elasticsearch, timed out after 1 minute")
	}

	storageManager := &StorageManager{
		client:       esclient,
		defaultIndex: conf.Index,
	}
	storageManager.createIndex(conf.Index)
	return storageManager, nil
}

// Save new documents
func (sm *StorageManager) Save(data string) bool {

	_, err := sm.client.Index().
		Index(sm.defaultIndex).
		BodyJson(data).
		Do(context.Background())

	if err != nil {
		log.WithFields(log.Fields{
			"index": sm.defaultIndex,
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
	executionIDQuery := elastic.NewMatchQuery("ExecutionID", executionID)
	eventTypeQuery := elastic.NewMatchQuery("EventType", "service_status")

	log.WithFields(log.Fields{
		"execution_id": executionIDQuery,
		"event_type":   eventTypeQuery,
	}).Debug("Going to get get summary with the following fields")

	searchResult, err := sm.client.Search().
		Query(elastic.NewBoolQuery().Must(eventTypeQuery, executionIDQuery)).
		Pretty(true).
		Size(100).
		Do(context.Background())

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
	dynamicMatchQuery = append(dynamicMatchQuery, elastic.NewMatchQuery("ExecutionID", executionID))
	dynamicMatchQuery = append(dynamicMatchQuery, elastic.NewMatchQuery("EventType", "resource_detected"))

	searchResult, err := sm.client.Search().
		Query(elastic.NewBoolQuery().Must(dynamicMatchQuery...)).
		Aggregation("sum", elastic.NewSumAggregation().Field("Data.PricePerMonth")).
		Size(0).Do(context.Background())

	if nil != err {
		log.WithError(err).WithFields(log.Fields{
			"filters":      filters,
			"milliseconds": searchResult.TookInMillis,
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
		Filters(elastic.NewBoolQuery().Filter(elastic.NewBoolQuery().Should(elastic.NewMatchQuery("EventType", "service_status")))).
		SubAggregation("ExecutionIDDesc", elastic.NewTermsAggregation().Field("ExecutionID").Size(queryLimit).Order("MaxEventTime", false).
			SubAggregation("MaxEventTime", elastic.NewMaxAggregation().Field("EventTime")))).
		Do(context.Background())

	if nil != err {
		log.WithError(err).WithFields(log.Fields{
			"milliseconds": searchResult.TookInMillis,
		}).Error("error when trying to get executions collectors")
		return executions, nil
	}

	resp, ok := searchResult.Aggregations.Terms("orderedExecutionID")
	if !ok {
		log.Error("orderedExecutionID field term does not exist")
		return executions, nil
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
			data := strings.Split(executionID, "_")

			// Remove the last element of Data which is the timestamp and leave all the others elements
			// Which construct the executionName
			executionName := strings.Join(data[:len(data)-1], "_")

			// Always take the last element which is the timestamp of the collector's run
			collectorExecutionTime, err := strconv.ParseInt(data[len(data)-1], 10, 64)
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
	componentQ := elastic.NewMatchQuery("EventType", "resource_detected")
	deploymentQ := elastic.NewMatchQuery("ExecutionID", executionID)
	ResourceNameQ := elastic.NewMatchQuery("ResourceName", resourceType)
	generalQ := elastic.NewBoolQuery()
	generalQ = generalQ.Must(componentQ).Must(deploymentQ).Must(ResourceNameQ).Must(dynamicMatchQuery...)
	searchResult, err := sm.client.Search().
		Query(generalQ).
		Pretty(true).
		Size(100).
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
func (sm *StorageManager) GetResourceTrends(resourceType string, filters map[string]string) ([]storage.ExecutionCost, error) {
	// Per resource trends, filters should take care of granularity (per resource: Data.ResourceID, Data.Region, Data.Metric -> Data.PricePerMonth)
	var resources []storage.ExecutionCost
	var mustNotQuery []elastic.Query

	// Must
	mustQuery := sm.getDynamicMatchQuery(filters, "and")
	mustQuery = append(mustQuery, elastic.NewMatchQuery("ResourceName", resourceType).Operator("and"))

	// Unsupported Types - MustNot
	mustNotQuery = append(mustNotQuery, elastic.NewMatchQuery("EventType", "service_status").Operator("and"))
	mustNotQuery = append(mustNotQuery, elastic.NewMatchQuery("ResourceName", "aws_iam_users").Operator("and"))
	mustNotQuery = append(mustNotQuery, elastic.NewMatchQuery("ResourceName", "aws_elastic_ip").Operator("and"))
	mustNotQuery = append(mustNotQuery, elastic.NewMatchQuery("ResourceName", "aws_lambda").Operator("and"))
	mustNotQuery = append(mustNotQuery, elastic.NewMatchQuery("ResourceName", "aws_ec2_volume").Operator("and"))

	queryBuilder := elastic.NewBoolQuery().MustNot(mustNotQuery...).Must(mustQuery...)
	searchResult, err := sm.client.Search().
		Query(queryBuilder).
		Pretty(true).
		Size(100).
		SortBy(elastic.NewFieldSort("Timestamp").Desc()).
		Aggregation("executions", elastic.NewTermsAggregation().Field("ExecutionID").OrderByKeyAsc(). // Aggregate by ExecutionID
														SubAggregation("monthly-cost", elastic.NewSumAggregation().Field("Data.PricePerMonth"))). // Sub aggregate and sum by Data.PricePerMonth per bucket
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("elasticsearch query error")
		return resources, err
	}

	executions, found := searchResult.Aggregations.Terms("executions")
	if found {
		for _, ppm := range executions.Buckets {
			executionId := ppm.Key.(string)
			monthlyAgg, _ := ppm.Aggregations.Sum("monthly-cost")

			// Extract the timestamp from the ExecutionID
			executionText := strings.Split(executionId, "_")
			timestamp, err := strconv.Atoi(executionText[1])
			var timestampInt int
			if err == nil {
				timestampInt = timestamp
			}

			resources = append(resources, storage.ExecutionCost{
				ExecutionID:        executionId,
				ExtractedTimestamp: timestampInt,
				CostSum:            *monthlyAgg.Value,
			})
		}
	}

	return resources, nil
}

// GetExecutionTags will return the tags according to a given executionID
func (sm *StorageManager) GetExecutionTags(executionID string) (map[string][]string, error) {

	tags := map[string][]string{}
	eventTypeMatchQuery := elastic.NewMatchQuery("EventType", "resource_detected")
	executionIDMatchQuery := elastic.NewMatchQuery("ExecutionID", executionID)
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
func (sm *StorageManager) createIndex(index string) {

	exists, err := sm.client.IndexExists(index).Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"index": index,
		}).WithError(err).Error("Error when trying to check if elasticsearch exists")
		return
	}
	if exists {
		log.WithField("index", index).Info("index already exists")
		return
	}

	ctx := context.Background()
	_, err = sm.client.CreateIndex(index).BodyString(indexMapping).Do(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"index": index,
		}).WithError(err).Error("Error when trying to create elasticsearch index")
	}

	log.WithField("index", index).Info("index created successfully")

}
