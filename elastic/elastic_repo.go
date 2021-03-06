package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/tidwall/gjson"
)

type ElasticRepository struct {
	es            *elasticsearch.Client
	numWorkers    int
	flushBytes    int
	flushInterval time.Duration
}

func NewDefaultClient() (*ElasticRepository, error) {
	es, err := NewClient("http://localhost:9200", "", "", "")
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

func constructQuery(q string, size int) *strings.Reader {

	// Build a query string from string passed to function
	var query = `{"query": {`

	// Concatenate query string with string passed to method call
	query = query + q

	// Use the strconv.Itoa() method to convert int to string
	query = query + `}, "size": ` + strconv.Itoa(size) + `}`
	fmt.Println("\nquery:", query)

	// Check for JSON errors
	isValid := json.Valid([]byte(query)) // returns bool

	// Default query is "{}" if JSON is invalid
	if isValid == false {
		fmt.Println("constructQuery() ERROR: query string not valid:", query)
		fmt.Println("Using default match_all query")
		query = "{}"
	} else {
		fmt.Println("constructQuery() valid JSON:", isValid)
	}
	// Build a new string from JSON query
	var b strings.Builder
	b.WriteString(query)

	// Instantiate a *strings.Reader object from string
	read := strings.NewReader(b.String())

	// Return a *strings.Reader object
	return read
}

func NewClient(url, user, pwd, caPath string) (*ElasticRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			url,
		},
	}

	if caPath != "" {
		cert, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		cfg.CACert = cert
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

func (repo *ElasticRepository) extractResults(res *esapi.Response, handler func(string), batchNum int) string {
	// Handle the first batch of data and extract the scrollID
	//
	json := read(res.Body)
	res.Body.Close()

	// Extract the scrollID from response
	scrollID := gjson.Get(json, "_scroll_id").String()

	// Extract the search results
	hits := gjson.Get(json, "hits.hits")

	if len(hits.Array()) < 1 {
		log.Println("Finished scrolling")
		return ""
	} else {
		log.Println("Batch   ", batchNum)
		log.Println("ScrollID", scrollID)
		log.Println("IDs     ", gjson.Get(hits.Raw, "#._id").Array())
		results := gjson.Get(hits.Raw, "#._source.url").Array()
		for _, result := range results {
			handler(result.String())
		}
		log.Println(strings.Repeat("-", 80))
	}
	return scrollID
}

func (repo *ElasticRepository) FindAllUrls(index string, limit int, handler func(string)) error {
	repo.Refresh(index)
	log.Println("Scrolling the index...")
	log.Println(strings.Repeat("-", 80))
	res, err := repo.es.Search(
		repo.es.Search.WithIndex(index),
		repo.es.Search.WithSort("_doc"),
		repo.es.Search.WithSize(limit),
		repo.es.Search.WithScroll(time.Minute),
	)
	if err != nil {
		return err
	}

	var batchNum int

	scrollID := repo.extractResults(res, handler, batchNum)

	for scrollID != "" {
		batchNum++

		// Perform the scroll request and pass the scrollID and scroll duration
		//
		res, err := repo.es.Scroll(repo.es.Scroll.WithScrollID(scrollID), repo.es.Scroll.WithScroll(time.Minute))
		if err != nil {
			return err
		}
		if res.IsError() {
			return fmt.Errorf("error response: %s", res)
		}

		scrollID = repo.extractResults(res, handler, batchNum)

	}

	return nil
}

func (repo *ElasticRepository) ExistUrl(index string, url string) (bool, error) {

	var (
		r map[string]interface{}
	)

	var query = `
	{
		"query":
		{
			"match":
			{
				"url":"$url"
			}
		}
	}`
	query = strings.Replace(query, "$url", url, 2)

	log.Println("Correct query: ", query)

	repo.Refresh(index)
	res, err := repo.es.Count(
		repo.es.Count.WithIndex(index),
		repo.es.Count.WithBody(strings.NewReader(query)),
	)

	if err != nil {
		log.Fatal("Error during execute the request")
	}

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
		return false, err
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	log.Println("Received response: ", r)

	//WIP: Validate the result
	if int(r["_shards"].(map[string]interface{})["successful"].(float64)) >= 1 {
		return true, nil
	} else {
		return false, nil
	}

}

func (repo *ElasticRepository) Refresh(index string) error {
	r := esapi.IndicesRefreshRequest{
		Index: []string{index},
	}
	_, err := r.Do(context.Background(), repo.es)
	return err
}

func (repo *ElasticRepository) Count(index string) (int64, error) {
	r := esapi.CountRequest{
		Index: []string{index},
	}
	res, err := r.Do(context.Background(), repo.es)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	json := read(res.Body)
	return gjson.Get(json, "count").Int(), nil
}

func read(r io.Reader) string {
	var b bytes.Buffer
	b.ReadFrom(r)
	return b.String()
}
