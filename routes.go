package brain

func (s *server) routes() {
	s.router.HandleFunc("/api/v1/indicator/url", s.insertUrl())
}