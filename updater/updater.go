package updater

import (
	"github.com/Ringloop/pisec/elastic"
	"github.com/bits-and-blooms/bloom/v3"
)

type Updater struct {
	elasticRepo *elastic.ElasticRepository
}

func NewUpdater(es *elastic.ElasticRepository) (*Updater, error) {
	return &Updater{es}, nil
}

func (u *Updater) DownloadIndicators() ([]byte, error) {
	bloomFilter := bloom.NewWithEstimates(1000000, 0.01)

	handler := func(url string) {
		bloomFilter.AddString(url)
	}

	err := u.elasticRepo.FindAllUrls("denylist", 1000, handler)
	if err != nil {
		return nil, err
	}

	json, err := bloomFilter.MarshalJSON()
	return json, err
}
