package brain

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGreet(t *testing.T) {

	server := NewTestServer()
	server.routes()

	test := UrlsBulkRequest{
		Source: "test-source",
		Indicators: []Indicator{
			{"google.com", "", 99},
			{"evil.com", "127.0.0.1", 50},
		},
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(test)
	if err != nil {
		panic("error in encoding test req")
	}
	_, _ = http.NewRequest(http.MethodPost, "/api/v1/indicator/url", &buf)

	req := httptest.NewRequest("POST", "/api/v1/indicator/url", &buf)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

}
