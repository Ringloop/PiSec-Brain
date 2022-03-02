package brain

type UrlsBulkRequest struct {
	Indicators []Indicator `json:"indicators"`
	Source     string
}

type Indicator struct {
	Url         string `json:"url"`
	Ip          string `json:"ip"`
	Reliability int    `json:"reliability"`
}

type ElasticIndicator struct {
	Url         string   `json:"url"`
	Ip          []string `json:"ip"`
	Source      string   `json:"source"`
	Reliability int      `json:"reliability"`
	//TODO dates
}

type OkResponse struct {
	Status string `json:"status"`
}
