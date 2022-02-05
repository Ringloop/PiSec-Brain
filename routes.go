package brain

func (s *server) routes() {
	s.router.HandleFunc("/api/v1/ndicator/url", s.insertUrl())
}
