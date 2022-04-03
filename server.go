package brain

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Ringloop/pisec/elastic"
	"github.com/Ringloop/pisec/updater"
	"github.com/gorilla/mux"
)

// -------------------   server.go
type server struct {
	denylist *Denylist
	updater  *updater.Updater
	router   *mux.Router
}

func NewTestServer() *server {
	es, err := elastic.NewDefaultClient()
	if err != nil {
		log.Default().Fatal(err) //todo error handling
	}
	denyList, err := NewDenylist(es)
	if err != nil {
		log.Default().Fatal(err) //todo error handling
	}

	updater, err := updater.NewUpdater(es)
	if err != nil {
		log.Default().Fatal(err)
	}

	s := &server{denyList, updater, mux.NewRouter()}
	s.routes()
	return s
}

// NewServer creates a server with router and does all things from here
func NewBrainServer() {
	es, err := elastic.NewEnvConfigClient()
	if err != nil {
		log.Default().Fatal(err) //todo error handling
	}
	denyList, err := NewDenylist(es)
	if err != nil {
		log.Default().Fatal(err)
	}
	updater, err := updater.NewUpdater(es)
	if err != nil {
		log.Default().Fatal(err)
	}

	s := &server{denyList, updater, mux.NewRouter()}
	s.routes()
	log.Fatalln(http.ListenAndServe(":8080", s.router))
}

func (s *server) insertUrl() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("serving request...")

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

		ok, err := json.Marshal(OkResponse{"Ok"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.denylist.AddUrls(&indicators)

		w.Header().Set("Content-Type", "application/json")
		w.Write(ok)
	}
}

func (s *server) downloadUpdates() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("serving request...")

		if r.Method != "GET" {
			http.Error(w, "method not supported", http.StatusBadRequest)
			return
		}

		json, err := s.updater.DownloadIndicators()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
}
