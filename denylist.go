package brain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	b64 "encoding/base64"

	"github.com/Ringloop/pisec/elastic"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

type Denylist struct {
	elasticRepo *elastic.ElasticRepository
}

func NewDenylist() (*Denylist, error) {
	es, err := elastic.NewClient("https://localhost:9200", "elastic", "integration-test") //todo config here...
	if err != nil {
		fmt.Println("cannot connect to es")
		return nil, err
	}

	fmt.Println("creating index mapping")
	err = es.CreateIndex("denylist")
	if err != nil {
		fmt.Println("cannot create mapping!")
		return nil, err
	}

	return &Denylist{es}, nil
}

func (denyList *Denylist) AddUrls(indicators *UrlsBulkRequest) error {

	elasticBulk, err := denyList.elasticRepo.GetBulkIndexer("denylist")
	if err != nil {
		return err
	}
	defer elasticBulk.Close(context.Background())

	for _, ind := range indicators.Indicators {

		toIndex := &ElasticIndicator{Date: makeTimestamp()}
		toIndex.Source = indicators.Source
		toIndex.Url = ind.Url

		if ind.Ip == "" {
			ips, err := net.LookupIP(ind.Url)
			if err != nil {
				return err
			}
			//toIndex.Ip = make([]string, 1)
			for _, resolvedIp := range ips {
				if len(resolvedIp) > 0 {
					fmt.Println("adding")
					fmt.Println(resolvedIp)
					toIndex.Ip = append(toIndex.Ip, resolvedIp.String())
				}
			}

		} else {
			//toIndex.Ip = make([]string, 1)
			toIndex.Ip = append(toIndex.Ip, ind.Ip)
		}

		fmt.Println(toIndex)
		fmt.Println(len(toIndex.Ip))
		documentToSend, err := json.Marshal(toIndex)
		if err != nil {
			return err
		}

		bulkItem := esutil.BulkIndexerItem{
			DocumentID: b64.StdEncoding.EncodeToString([]byte(toIndex.Url)),
			Action:     "index",
			Body:       bytes.NewBuffer(documentToSend),
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					log.Printf("ERROR: %s", err)
				} else {
					log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
				}
			},
		}

		err = elasticBulk.Add(context.Background(), bulkItem)
		if err != nil {
			return err
		}

		fmt.Println("done...")
	}

	return nil
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
