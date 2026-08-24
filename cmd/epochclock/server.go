package main

import (
	"net/http"
	"time"

	"epochclock/internal/alloc"
	"epochclock/internal/clock"
	"epochclock/internal/lease"
	"epochclock/internal/metric"
	"epochclock/internal/node"
	"epochclock/internal/persist"
)

// ServerDeps 聚合 HTTP 服务依赖的全部组件。
type ServerDeps struct {
	Allocator *alloc.Allocator
	Leases    *lease.Manager
	Nodes     *node.Registry
	HLC       *clock.HLC
	Metrics   *metric.Metrics
	Persister persist.Persister
	StartedAt time.Time
}

// Server 是 EpochClock 的 HTTP 服务。
type Server struct {
	deps ServerDeps
	mux  *http.ServeMux
}

// NewServer 创建服务并注册路由。
func NewServer(deps ServerDeps) *Server {
	server := &Server{
		deps: deps,
		mux:  http.NewServeMux(),
	}
	server.routes()
	return server
}

// Handler 返回可供 http.Server 使用的处理器。
func (s *Server) Handler() http.Handler {
	return s.mux
}
