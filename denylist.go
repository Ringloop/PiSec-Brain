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

	"github.com/Ringloop/pisec/cache"
	"github.com/Ringloop/pisec/elastic"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

type Denylist struct {
	elasticRepo *elastic.ElasticRepository
	redisRepo   *cache.RedisRepository
}

func NewDenylist(es *elastic.ElasticRepository, redisRepo *cache.RedisRepository) (*Denylist, error) {
	fmt.Println("pisec brain started, creating index mapping...")
	err := es.CreateIndex("denylist")
	if err != nil {
		fmt.Println("cannot create mapping!")
		return nil, err
	}
	return &Denylist{es, redisRepo}, nil
}

func (denyList *Denylist) CheckUrl(url string) (bool, error) {
	found, err := denyList.elasticRepo.ExistUrl("denylist", url)
	if err != nil {
		return false, err
	}

	return found, err
}

func (denyList *Denylist) DownloadIndicators() ([]byte, error) {
	bloomFilter := bloom.NewWithEstimates(1000000, 0.01)

	denyList.redisRepo.FindAllDenyList()

	json, err := bloomFilter.MarshalJSON()
	return json, err
}

func (denyList *Denylist) AddUrls(indicators *UrlsBulkRequest) error {

	elasticBulk, err := denyList.elasticRepo.GetBulkIndexer("denylist")
	if err != nil {
		return err
	}
	defer elasticBulk.Close(context.Background())

	bloomFilter := bloom.NewWithEstimates(1000000, 0.01)

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

		bloomFilter.AddString(ind.Url)

	}

	mergeHadler := func(oldValue string) (string, error) {
		if oldValue == "" {
			return bloomFilterToJson(bloomFilter)
		}

		oldBloomFilter := bloom.NewWithEstimates(1000000, 0.01)
		if err := oldBloomFilter.UnmarshalJSON([]byte(oldValue)); err != nil {
			return "", err
		}

		oldBloomFilter.Merge(bloomFilter)
		return bloomFilterToJson(oldBloomFilter)

	}
	denyList.redisRepo.AddDeny(indicators.Source, getDayOfTheMonth(), mergeHadler)

	return nil
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func getDayOfTheMonth() int {
	_, _, day := time.Now().Date()
	return day
}

func bloomFilterToJson(filter *bloom.BloomFilter) (string, error) {
	json, err := filter.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(json), nil
}
