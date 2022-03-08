package elastic

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

type ElasticRepository struct {
	es            *elasticsearch.Client
	numWorkers    int
	flushBytes    int
	flushInterval time.Duration
}

func NewDefaultClient() (*ElasticRepository, error) {
	es, err := NewClient(
		"https://localhost:9200",
		"elastic", "integration-test",
		"./pisec-brain-docker/certs/ca/ca.crt")
	if err != nil {
		return nil, err
	}
	return es, err
}

func NewEnvConfigClient() (*ElasticRepository, error) {
	es, err := NewClient(
		os.Getenv("ES_HOST"),
		os.Getenv("ES_USER"),
		os.Getenv("ES_PWD"),
		os.Getenv("ES_CA_CERT"))
	if err != nil {
		return nil, err
	}
	return es, err
}

func NewClient(url, user, pwd, caPath string) (*ElasticRepository, error) {
	cert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			url,
		},
		CACert: cert,
	}

	if user != "" {
		cfg.Username = user
		cfg.Password = pwd
	}

	if es, err := elasticsearch.NewClient(cfg); err != nil {
		return &ElasticRepository{}, err
	} else {
		return &ElasticRepository{
			es:            es,
			numWorkers:    1,
			flushBytes:    100000,
			flushInterval: 30 * time.Second}, nil
	}
}

func (repo *ElasticRepository) CreateIndex(index string) error {
	allowIndexNotExists := true

	indexReq := &esapi.IndicesCreateRequest{
		Index: index,
	}
	resp, err := indexReq.Do(context.Background(), repo.es)
	if err != nil {
		return err
	}
	fmt.Println(resp.String())

	mappingReq := &esapi.IndicesPutMappingRequest{
		Index: []string{index},
		Body: strings.NewReader(`
		{
			
			  "properties": {
				"date": {
				  "type": "date" 
				},
				"ip": {
					"type": "ip"
				},
				"source": {
					"type": "keyword"
				},
				"url": {
					"type": "keyword"
				},
				"reliability" : {
					"type" : "long"
				}
			  }
			
		}
		
		`),
		AllowNoIndices: &allowIndexNotExists,
	}

	resp, err = mappingReq.Do(context.Background(), repo.es)
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.String())
	return err
}

func (repo *ElasticRepository) GetBulkIndexer(index string) (esutil.BulkIndexer, error) {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         index,
		Client:        repo.es,
		NumWorkers:    repo.numWorkers,
		FlushBytes:    repo.flushBytes,
		FlushInterval: repo.flushInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting bulkIndexer: %s", err)
	}
	return bi, nil
}

func (repo *ElasticRepository) Delete(index string) error {
	_, err := repo.es.Indices.Delete([]string{index})
	return err
}
