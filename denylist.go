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

func NewDenylist(es *elastic.ElasticRepository) (*Denylist, error) {
	fmt.Println("pisec brain started, creating index mapping...")
	err := es.CreateIndex("denylist")
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
		toIndex.Reliability = ind.Reliability

		if ind.Ip == "" {
			ips, err := net.LookupIP(ind.Url)
			if err != nil {
				return err
			}

			for _, resolvedIp := range ips {
				if len(resolvedIp) > 0 {
					toIndex.Ip = append(toIndex.Ip, resolvedIp.String())
				}
			}

		} else {
			//toIndex.Ip = make([]string, 1)
			toIndex.Ip = append(toIndex.Ip, ind.Ip)
		}

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

		
	}

	return nil
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
