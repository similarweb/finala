package elasticsearch

// OrderedExecutionIDs will be the unmarshal response for ElasticSearch query  GetExecutions function
type orderedExecutionIDs struct {
	Buckets []struct {
		Key string `json:"key"`
	} `json:"buckets"`
}

// TagsData represents the response for ElasticSearch query of Get Execution tags
type TagsData struct {
	Data struct {
		Tag map[string]string `json:"Tag"`
	} `json:"Data"`
}
