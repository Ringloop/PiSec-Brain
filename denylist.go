package brain

import "net"

type Denylist struct {
}

func (*Denylist) AddUrls(indicators *UrlsBulkRequest) error {
	for _, ind := range indicators.Indicators {
		if ind.Ip == "" {
			ips, err := net.LookupIP(ind.Url)
			if err != nil {
				return err
			}
			ind.Ip = ips
		}
	}

	return nil
}
