package brain

type UrlsBulkRequest struct {
	Indicators []Indicator `json:"indicators"`
	Source     string
}

type Indicator struct {
	Url string `json:"url"`
	Ip  string `json:"ip"`
}

type OkResponse struct {
	Status string `json:"status"`
}
