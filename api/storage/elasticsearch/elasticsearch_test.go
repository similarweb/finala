package elasticsearch

import (
	"encoding/json"
	"finala/api/storage/elasticsearch/testutils"
	"fmt"
	"net/http"
	"testing"

	elastic "github.com/olivere/elastic/v7"
)

func TestESConnection(t *testing.T) {

	t.Run("successfull init storage manager", func(t *testing.T) {
		_, config := testutils.NewESMock(prefixIndexName, true)
		_, err := NewStorageManager(config)

		if err != nil {
			t.Fatalf("unexpected error, got %v expected nil", err)
		}
	})

	t.Run("error init storage manager", func(t *testing.T) {
		_, config := testutils.NewESMock(prefixIndexName, false)
		_, err := NewStorageManager(config)

		if err.Error() != "could not create index" {
			t.Fatalf("unexpected error, got %v expected error: %s", err, "could not create index")
		}
	})

}

func TestSave(t *testing.T) {

	t.Run("save successfull", func(t *testing.T) {
		mockClient, config := testutils.NewESMock(prefixIndexName, true)

		mockClient.Router.HandleFunc(fmt.Sprintf("/%s/_doc/", mockClient.DefaultIndex), func(resp http.ResponseWriter, req *http.Request) {
			testutils.JSONResponse(resp, http.StatusCreated, elastic.IndexResponse{Index: "test"})
		})

		es, err := NewStorageManager(config)

		if err != nil {
			t.Fatalf("unexpected error, got %v expected nil", err)
		}

		if !es.Save(string(testutils.GetDummyDoc("aws_resource_name", nil))) {
			t.Fatal("unexpected save data, got false expected true")
		}
	})

	t.Run("save failed", func(t *testing.T) {

		mockClient, config := testutils.NewESMock(prefixIndexName, true)

		mockClient.Router.HandleFunc(fmt.Sprintf("/%s/_doc/", mockClient.DefaultIndex), func(resp http.ResponseWriter, req *http.Request) {
			testutils.JSONResponse(resp, http.StatusBadRequest, elastic.IndexResponse{Index: "test"})
		})

		es, err := NewStorageManager(config)

		if err != nil {
			t.Fatalf("unexpected error, got %v expected nil", err)
		}

		if es.Save(string(testutils.GetDummyDoc("aws_resource_name", nil))) {
			t.Fatal("unexpected save data, got true expected false")

		}
	})

}

func TestGetDynamicMatchQuery(t *testing.T) {

	_, config := testutils.NewESMock(prefixIndexName, true)

	es, err := NewStorageManager(config)
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	filters := map[string]string{
		"test1": "foo",
	}

	t.Run("default", func(t *testing.T) {

		query := es.getDynamicMatchQuery(filters, "or")

		if len(query) != len(filters) {
			t.Fatalf("handler query response: got %d want %d", len(query), len(filters))
		}

		firstQuery, err := query[0].Source()
		if err != nil {
			t.Fatalf("unexpected error, got %v expected nil", err)
		}
		queryJSON, err := json.Marshal(firstQuery)
		if err != nil {
			t.Fatalf("unexpected error, got %v expected nil", err)
		}

		expectedFilter := `{"match":{"test1":{"minimum_should_match":"100%","query":"foo"}}}`
		if string(queryJSON) != expectedFilter {
			t.Fatalf("unexpected query filter: got %s want %s", string(queryJSON), expectedFilter)
		}

	})

	t.Run("with AND", func(t *testing.T) {

		query := es.getDynamicMatchQuery(filters, "and")

		if len(query) != len(filters) {
			t.Fatalf("handler query response: got %d want %d", len(query), len(filters))
		}

		firstQuery, err := query[0].Source()
		if err != nil {
			t.Fatalf("unexpected error, got %v expected nil", err)
		}
		queryJSON, err := json.Marshal(firstQuery)
		if err != nil {
			t.Fatalf("unexpected error, got %v expected nil", err)
		}

		expectedFilter := `{"match":{"test1":{"minimum_should_match":"100%","operator":"and","query":"foo"}}}`
		if string(queryJSON) != expectedFilter {
			t.Fatalf("unexpected query filter: got %s want %s", string(queryJSON), expectedFilter)
		}

	})

	t.Run("multiple", func(t *testing.T) {

		multipleFilters := map[string]string{
			"test1": "foo",
			"test2": "foo",
		}
		query := es.getDynamicMatchQuery(multipleFilters, "and")

		if len(query) != len(multipleFilters) {
			t.Fatalf("handler query response: got %d want %d", len(query), len(multipleFilters))
		}

	})

}

func TestGetExecutions(t *testing.T) {

	// each different queryLimit value will be a different elasticsarch response.
	// 1 - returns valid executions response
	// 2 - returns invalid term query response
	// 3 - returns invalid aggregation key
	// 4 - returns invalid statuscode response
	testCases := []struct {
		name          string
		queryLimit    int
		responseCount int
		ErrorMessage  error
	}{
		{"valid response", 1, 2, nil},
		{"invalid terms", 2, 0, ErrAggregationTermNotFound},
		{"invalid aggregations key", 3, 0, nil},
		{"invalid es response", 4, 0, ErrInvalidQuery},
	}

	mockClient, config := testutils.NewESMock(prefixIndexName, true)

	mockClient.Router.HandleFunc("/_search", func(resp http.ResponseWriter, req *http.Request) {

		switch testutils.GetPostParams(req) {
		case `{"aggregations":{"orderedExecutionID":{"aggregations":{"ExecutionIDDesc":{"aggregations":{"MaxEventTime":{"max":{"field":"EventTime"}}},"terms":{"field":"ExecutionID","order":[{"MaxEventTime":"desc"}],"size":1}}},"filters":{"filters":[{"bool":{"filter":{"bool":{"should":{"term":{"EventType":"service_status"}}}}}}]}}}}`:
			testutils.JSONResponse(resp, http.StatusOK, elastic.SearchResult{Aggregations: map[string]json.RawMessage{
				"orderedExecutionID": testutils.LoadResponse("executions/aggregations/default"),
			}})
		case `{"aggregations":{"orderedExecutionID":{"aggregations":{"ExecutionIDDesc":{"aggregations":{"MaxEventTime":{"max":{"field":"EventTime"}}},"terms":{"field":"ExecutionID","order":[{"MaxEventTime":"desc"}],"size":2}}},"filters":{"filters":[{"bool":{"filter":{"bool":{"should":{"term":{"EventType":"service_status"}}}}}}]}}}}`:
			testutils.JSONResponse(resp, http.StatusOK, elastic.SearchResult{Aggregations: map[string]json.RawMessage{
				"invalid-key": testutils.LoadResponse("executions/aggregations/default"),
			}})
		case `{"aggregations":{"orderedExecutionID":{"aggregations":{"ExecutionIDDesc":{"aggregations":{"MaxEventTime":{"max":{"field":"EventTime"}}},"terms":{"field":"ExecutionID","order":[{"MaxEventTime":"desc"}],"size":3}}},"filters":{"filters":[{"bool":{"filter":{"bool":{"should":{"term":{"EventType":"service_status"}}}}}}]}}}}`:
			testutils.JSONResponse(resp, http.StatusOK, elastic.SearchResult{Aggregations: map[string]json.RawMessage{
				"orderedExecutionID": testutils.LoadResponse("executions/aggregations/invalid-aggregations-key"),
			}})
		case `{"aggregations":{"orderedExecutionID":{"aggregations":{"ExecutionIDDesc":{"aggregations":{"MaxEventTime":{"max":{"field":"EventTime"}}},"terms":{"field":"ExecutionID","order":[{"MaxEventTime":"desc"}],"size":4}}},"filters":{"filters":[{"bool":{"filter":{"bool":{"should":{"term":{"EventType":"service_status"}}}}}}]}}}}`:
			testutils.JSONResponse(resp, http.StatusBadRequest, elastic.SearchResult{Aggregations: map[string]json.RawMessage{}})
		default:
			t.Fatalf("unexpected request params")
		}

	})

	es, err := NewStorageManager(config)
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			response, err := es.GetExecutions(test.queryLimit)

			if err != test.ErrorMessage {
				t.Fatalf("unexpected error, got %v expected %v", err, test.ErrorMessage)
			}

			if len(response) != test.responseCount {
				t.Fatalf("handler query response: got %d want %d", len(response), test.responseCount)
			}

		})
	}
}

func TestGetSummary(t *testing.T) {

	mockClient, config := testutils.NewESMock(prefixIndexName, true)

	mockClient.Router.HandleFunc("/_search", func(resp http.ResponseWriter, req *http.Request) {

		response := elastic.SearchResult{}

		switch testutils.GetPostParams(req) {
		case `{"query":{"bool":{"must":[{"term":{"EventType":"service_status"}},{"term":{"ExecutionID":""}}]}},"size":0}`:
			response.Hits = &elastic.SearchHits{TotalHits: &elastic.TotalHits{Value: 1}}
		case `{"query":{"bool":{"must":[{"term":{"EventType":"service_status"}},{"term":{"ExecutionID":""}}]}},"size":1}`:
			response.Hits = &elastic.SearchHits{
				TotalHits: &elastic.TotalHits{Value: 1},
				Hits: []*elastic.SearchHit{
					{Source: testutils.GetDummyDoc("aws_resource_name", nil)},
				},
			}
		case `{"aggregations":{"sum":{"sum":{"field":"Data.PricePerMonth"}}},"query":{"bool":{"must":[{"match":{"ResourceName":{"minimum_should_match":"100%","query":"aws_resource_name"}}},{"term":{"ExecutionID":""}},{"term":{"EventType":"resource_detected"}}]}},"size":0}`:
			response.Aggregations = map[string]json.RawMessage{"sum": []byte(`{"value": 36.5}`)}
			response.Hits = &elastic.SearchHits{TotalHits: &elastic.TotalHits{Value: 1}}
		default:
			t.Fatalf("unexpected request params")
		}

		testutils.JSONResponse(resp, http.StatusOK, response)

	})

	es, err := NewStorageManager(config)
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	summaryResponse, err := es.GetSummary("", map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	if len(summaryResponse) != 1 {
		t.Fatalf("unexpected summary response count: got %d want %d", len(summaryResponse), 2)
	}

	expectedResourceName := "aws_resource_name"
	data, ok := summaryResponse[expectedResourceName]
	if !ok {
		t.Fatalf("unexpected resource, got nil expected %s", expectedResourceName)
	}

	if data.ResourceName != expectedResourceName {
		t.Fatalf("unexpected resource name, got %s expected %s", data.ResourceName, expectedResourceName)
	}

	if data.ResourceCount != 1 {
		t.Fatalf("unexpected resource count, got %d expected %d", data.ResourceCount, 1)
	}

	if data.ResourceCount != 1 {
		t.Fatalf("unexpected resource count, got %d expected %d", data.ResourceCount, 1)
	}

	if data.TotalSpent != 36.5 {
		t.Fatalf("unexpected resource count, got %v expected %v", data.TotalSpent, 36.5)
	}

}

func TestGetResources(t *testing.T) {

	mockClient, config := testutils.NewESMock(prefixIndexName, true)

	mockClient.Router.HandleFunc("/_search", func(resp http.ResponseWriter, req *http.Request) {

		response := elastic.SearchResult{}

		switch testutils.GetPostParams(req) {
		case `{"query":{"bool":{"must":[{"term":{"EventType":"resource_detected"}},{"term":{"ExecutionID":"1234"}},{"term":{"ResourceName":"aws_resource_name"}},{"match":{"foo":{"minimum_should_match":"100%","query":"bar"}}}]}},"size":0}`:
			response.Hits = &elastic.SearchHits{
				TotalHits: &elastic.TotalHits{Value: 2},
			}
		case `{"query":{"bool":{"must":[{"term":{"EventType":"resource_detected"}},{"term":{"ExecutionID":"1234"}},{"term":{"ResourceName":"aws_resource_name"}},{"match":{"foo":{"minimum_should_match":"100%","query":"bar"}}}]}},"size":2}`:
			response.Hits = &elastic.SearchHits{
				TotalHits: &elastic.TotalHits{Value: 1},
				Hits: []*elastic.SearchHit{
					{Source: testutils.GetDummyDoc("aws_elb", nil)},
					{Source: testutils.GetDummyDoc("aws_elb", nil)},
					{Source: testutils.GetDummyDoc("aws_ec2", nil)},
				},
			}
		default:
			t.Fatalf("unexpected request params")
		}

		testutils.JSONResponse(resp, http.StatusOK, response)

	})

	es, err := NewStorageManager(config)
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	resourcesResponse, err := es.GetResources("aws_resource_name", "1234", map[string]string{"foo": "bar"})
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	if len(resourcesResponse) != 3 {
		t.Fatalf("unexpected resources response count: got %d want %d", len(resourcesResponse), 3)
	}

}

func TestGetResourceTrends(t *testing.T) {

	mockClient, config := testutils.NewESMock(prefixIndexName, true)

	mockClient.Router.HandleFunc("/_search", func(resp http.ResponseWriter, req *http.Request) {

		response := elastic.SearchResult{}

		switch testutils.GetPostParams(req) {
		case `{"query":{"bool":{"must":[{"match":{"foo":{"minimum_should_match":"100%","operator":"and","query":"bar"}}},{"term":{"ResourceName":"resource-name"}}],"must_not":[{"term":{"EventType":"service_status"}},{"term":{"ResourceName":"aws_iam_users"}},{"term":{"ResourceName":"aws_elastic_ip"}},{"term":{"ResourceName":"aws_lambda"}},{"term":{"ResourceName":"aws_ec2_volume"}}]}},"size":0}`:
			response.Hits = &elastic.SearchHits{
				TotalHits: &elastic.TotalHits{Value: 2},
			}
		case `{"aggregations":{"executions":{"aggregations":{"monthly-cost":{"sum":{"field":"Data.PricePerMonth"}}},"terms":{"field":"ExecutionID","order":[{"_key":"desc"}]}}},"query":{"bool":{"must":[{"match":{"foo":{"minimum_should_match":"100%","operator":"and","query":"bar"}}},{"term":{"ResourceName":"resource-name"}}],"must_not":[{"term":{"EventType":"service_status"}},{"term":{"ResourceName":"aws_iam_users"}},{"term":{"ResourceName":"aws_elastic_ip"}},{"term":{"ResourceName":"aws_lambda"}},{"term":{"ResourceName":"aws_ec2_volume"}}]}},"size":2,"sort":[{"Timestamp":{"order":"desc"}}]}`:
			response.Aggregations = map[string]json.RawMessage{
				"executions": testutils.LoadResponse("trends/buckets"),
			}
		default:
			t.Fatalf("unexpected request params")
		}

		testutils.JSONResponse(resp, http.StatusOK, response)

	})

	es, err := NewStorageManager(config)
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	trendResponse, err := es.GetResourceTrends("resource-name", map[string]string{"foo": "bar"}, 4)
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	if len(trendResponse) != 2 {
		t.Fatalf("unexpected trend response count: got %d want %d", len(trendResponse), 2)
	}

}

func TestGetExecutionTags(t *testing.T) {

	mockClient, config := testutils.NewESMock(prefixIndexName, true)

	mockClient.Router.HandleFunc("/_search", func(resp http.ResponseWriter, req *http.Request) {

		response := elastic.SearchResult{}

		switch testutils.GetPostParams(req) {
		case `{"query":{"bool":{"must":[{"term":{"EventType":"resource_detected"}},{"term":{"ExecutionID":"execution_id"}}]}},"size":0}`:
			response.Hits = &elastic.SearchHits{TotalHits: &elastic.TotalHits{Value: 10}}
		case `{"query":{"bool":{"must":[{"term":{"EventType":"resource_detected"}},{"term":{"ExecutionID":"execution_id"}}]}},"size":10}`:

			response.Hits = &elastic.SearchHits{
				TotalHits: &elastic.TotalHits{Value: 1},
				Hits: []*elastic.SearchHit{
					{Source: testutils.GetDummyDoc("1", map[string]interface{}{"tag": map[string]string{"a": "a", "b": "b", "c": "c"}})},
					{Source: testutils.GetDummyDoc("2", map[string]interface{}{"tag": map[string]string{"foo": "bar", "a": "a"}})},
				},
			}
		default:
			t.Fatalf("unexpected request params")
		}

		testutils.JSONResponse(resp, http.StatusOK, response)

	})

	es, err := NewStorageManager(config)
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	tags, err := es.GetExecutionTags("execution_id")
	if err != nil {
		t.Fatalf("unexpected error, got %v expected nil", err)
	}

	if len(tags) != 4 {
		t.Fatalf("unexpected tags response count: got %d want %d", len(tags), 4)
	}

}
