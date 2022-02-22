package brain

type UrlsBulkRequest struct {
	Indicators []Indicator `json:"indicators"`
	Source     string
}

type Indicator struct {
	Url string `json:"url"`
	Ip  string `json:"ip"`
}

type ElasticIndicator struct {
	Url    string
	Ip     []string
	Source string
	//TODO dates
}

type OkResponse struct {
	Status string `json:"status"`
}
