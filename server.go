package brain

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// -------------------   server.go
type server struct {
	denylist *Denylist
	router   *mux.Router
}

type okResponse struct {
	Status string `json:"status"`
}

// NewServer creates a server with router and does all things from here
func NewServer() {
	s := &server{&Denylist{}, mux.NewRouter()}
	s.routes()
	log.Fatalln(http.ListenAndServe(":8080", s.router))
}

func (s *server) insertUrl() http.HandlerFunc {

	type UrlsBulkRequest struct {
		Indicators []string `json:"indicators"`
		Source     string
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "method not supported", http.StatusBadRequest)
			return
		}

		var indicators UrlsBulkRequest

		err := json.NewDecoder(r.Body).Decode(&indicators)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ok, err := json.Marshal(okResponse{"Ok"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(ok)
	}
}
