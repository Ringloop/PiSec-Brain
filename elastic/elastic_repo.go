package elastic

import (
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

type ElasticRepository struct {
	es            *elasticsearch.Client
	numWorkers    int
	flushBytes    int
	flushInterval time.Duration
}

func NewDefaultClient() (*ElasticRepository, error) {
	if es, err := elasticsearch.NewDefaultClient(); err != nil {
		return &ElasticRepository{}, err
	} else {
		return &ElasticRepository{
			es:            es,
			numWorkers:    1,
			flushBytes:    100000,
			flushInterval: 30 * time.Second}, nil
	}
}

func NewClient(url, user, pwd string) (*ElasticRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			url,
		},
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
			numWorkers:    10,
			flushBytes:    100000,
			flushInterval: 30 * time.Second}, nil
	}
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
