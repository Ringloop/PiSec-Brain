package brain

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// -------------------   server.go
type server struct {
	denylist *Denylist
	router   *mux.Router
}

func NewTestServer() *server {
	s := &server{&Denylist{}, mux.NewRouter()}
	s.routes()
	return s
}

// NewServer creates a server with router and does all things from here
func NewBrainServer() {
	s := &server{&Denylist{}, mux.NewRouter()}
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
