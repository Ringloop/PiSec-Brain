package brain

import (
	"fmt"
	"net"
)

type Denylist struct {
}

func (*Denylist) AddUrls(indicators *UrlsBulkRequest) error {
	for _, ind := range indicators.Indicators {

		toIndex := &ElasticIndicator{}
		toIndex.Source = indicators.Source
		toIndex.Url = ind.Url

		if ind.Ip == "" {
			ips, err := net.LookupIP(ind.Url)
			if err != nil {
				return err
			}
			toIndex.Ip = make([]string, len(ips))
			for _, resolvedIp := range ips {
				toIndex.Ip = append(toIndex.Ip, resolvedIp.String())
			}

		} else {
			toIndex.Ip = make([]string, 1)
			toIndex.Ip = append(toIndex.Ip, ind.Ip)
		}

		fmt.Println(toIndex)
	}

	return nil
}
