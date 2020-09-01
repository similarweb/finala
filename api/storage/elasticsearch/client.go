package elasticsearch

import (
	"finala/api/config"
	"fmt"
	"strings"
	"time"

	elastic "github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

const (
	// connectionInterval defines the time duration to wait until the next connection retry
	connectionInterval = 5 * time.Second
	// connectionTimeout defines the maximum time duration until the API returns a connection error
	connectionTimeout = 60 * time.Second
)

// elasticSearchDescriptor is the ES root interface
type elasticSearchDescriptor interface {
	Index() *elastic.IndexService
	Search(indices ...string) *elastic.SearchService
	CreateIndex(name string) *elastic.IndicesCreateService
	IndexExists(indices ...string) *elastic.IndicesExistsService
}

// NewClient creates new elasticsearch client
func NewClient(conf config.ElasticsearchConfig) (*elastic.Client, error) {

	var esClient *elastic.Client

	c := make(chan int, 1)
	var err error
	go func() {

		for {
			esClient, err = getESClient(conf)
			if err == nil {
				break
			}
			log.WithFields(log.Fields{
				"endpoint": conf.Endpoints,
			}).WithError(err).Warn(fmt.Sprintf("could not initialize connection to elasticsearch, retrying in %v", connectionInterval))
			time.Sleep(connectionInterval)
		}
		c <- 1
	}()

	select {
	case <-c:
	case <-time.After(connectionTimeout):
		err = fmt.Errorf("could not connect elasticsearch, timed out after %v", connectionTimeout)
		log.WithError(err).Error("connection Error")
	}

	return esClient, err

}

// getESClient create new elasticsearch client
func getESClient(conf config.ElasticsearchConfig) (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL(strings.Join(conf.Endpoints, ",")),
		elastic.SetErrorLog(log.New()),
		// elastic.SetTraceLog(log.New()), // Uncomment for debugging ElasticSearch Queries
		elastic.SetBasicAuth(conf.Username, conf.Password),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(true))

	return client, err

}
