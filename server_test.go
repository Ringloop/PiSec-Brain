package brain

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ringloop/pisec/elastic"
)

func TestInsertAndDownload(t *testing.T) {
	//given
	repo, err := elastic.NewDefaultClient()
	if err != nil {
		panic(err)
	}
	repo.Delete("denylist")
	server := NewTestServer()
	server.routes()

	//when (sendind crawler request)
	test := UrlsBulkRequest{
		Source: "test-source",
		Indicators: []Indicator{
			{"google.com", "", 99},
			{"evil.com", "127.0.0.1", 50},
			{"evil.it", "127.0.0.1", 50},
		},
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(test)
	if err != nil {
		panic(err)
	}
	_, _ = http.NewRequest(http.MethodPost, "/api/v1/indicator/url", &buf)

	req := httptest.NewRequest("POST", "/api/v1/indicator/url", &buf)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	//then (the data is on elasticsearch)
	err = repo.FindAllUrls("denylist", 1, func(url string) {
		log.Println("found ", url)
	})
	if err != nil {
		panic(err)
	}
}
