package brain

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ringloop/pisec/elastic"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/stretchr/testify/require"
)

var testDenyElements = map[string]struct{}{"google.com": {}, "evil.com": {}, "evil.it": {}}
var testMissingDenyElements = map[string]struct{}{"google.it": {}, "evil.net": {}, "amNotEvil.com": {}}

var testIndicators = []Indicator{
	{"google.com", "", 99},
	{"evil.com", "127.0.0.1", 50},
	{"evil.it", "127.0.0.1", 50},
}

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
		Source:     "test-source",
		Indicators: testIndicators,
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
		require.Contains(t, testDenyElements, url)
		delete(testDenyElements, url)
	})
	if err != nil {
		panic(err)
	}

	require.Empty(t, testDenyElements)
}

func TestInsertAndDownloadIndicator(t *testing.T) {
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
		Source:     "test-source",
		Indicators: testIndicators,
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(test)
	if err != nil {
		panic(err)
	}
	_, _ = http.NewRequest(http.MethodPost, "/api/v1/indicator/url", &buf)

	insertReq := httptest.NewRequest("POST", "/api/v1/indicator/url", &buf)
	insertRec := httptest.NewRecorder()
	server.router.ServeHTTP(insertRec, insertReq)

	downloadIndicatorsReq := httptest.NewRequest("GET", "/api/v1/indicators", nil)
	downloadIndicatorsRec := httptest.NewRecorder()
	server.router.ServeHTTP(downloadIndicatorsRec, downloadIndicatorsReq)

	if want, got := http.StatusOK, downloadIndicatorsRec.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}

	res := downloadIndicatorsRec.Result()
	defer res.Body.Close()
	jsonRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	bFilter := bloom.NewWithEstimates(1000000, 0.01)
	bFilter.UnmarshalJSON(jsonRes)

	for key, _ := range testDenyElements {
		require.True(t, bFilter.TestString(key))
	}

	for key, _ := range testMissingDenyElements {
		require.False(t, bFilter.TestString(key))
	}

}
