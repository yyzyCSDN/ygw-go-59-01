package main

// routes 注册全部 HTTP 路由。
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /api/nodes/{id}", s.handleNode)
	s.mux.HandleFunc("POST /api/nodes/unregister", s.handleNodeUnregister)
	s.mux.HandleFunc("GET /api/ranges", s.handleRanges)
	s.mux.HandleFunc("GET /api/leases", s.handleLeases)
	s.mux.HandleFunc("POST /v1/sequences/issue", s.handleIssue)
	s.mux.HandleFunc("POST /v1/leases/failover", s.handleFailover)
	s.mux.HandleFunc("POST /v1/leases/expire", s.handleLeaseExpire)
	s.mux.HandleFunc("GET /monitor", s.handleMonitor)
	s.mux.HandleFunc("GET /", s.handleIndex)
}
