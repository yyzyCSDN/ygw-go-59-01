package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"epochclock/internal/alloc"
	"epochclock/internal/clock"
	"epochclock/internal/lease"
	"epochclock/internal/metric"
	"epochclock/internal/node"
	"epochclock/internal/persist"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	metrics := metric.New()
	nodes := node.NewRegistry()
	seed := nodes.Register("seed-node")
	hlc := clock.New(0, 0)
	leases := lease.NewManager(nodes, time.Now, metrics)
	pers, err := persist.NewWALPersister(t.TempDir(), 1024)
	if err != nil {
		t.Fatalf("open wal: %v", err)
	}
	t.Cleanup(func() { _ = pers.Close() })
	result, err := pers.Replay()
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	allocator := alloc.NewAllocator(pers, hlc, leases, metrics, clock.SystemWall(), 100)
	if err := allocator.Restore(result); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if _, err := leases.Grant(seed.ID, time.Minute); err != nil {
		t.Fatalf("grant: %v", err)
	}
	return NewServer(ServerDeps{
		Allocator: allocator,
		Leases:    leases,
		Nodes:     nodes,
		HLC:       hlc,
		Metrics:   metrics,
		Persister: pers,
		StartedAt: time.Now().UTC(),
	})
}

func TestHealthEndpoint(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("healthz status = %d", recorder.Code)
	}
	if recorder.Body.String() != "ok" {
		t.Fatalf("healthz body = %q", recorder.Body.String())
	}
}

func TestMonitorPage(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/monitor", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("monitor status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	if len(body) < 500 {
		t.Fatalf("monitor page too short: %d bytes", len(body))
	}
}

func TestStatusEndpoint(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d", recorder.Code)
	}
	var snapshot StatusSnapshot
	if err := json.Unmarshal(recorder.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if !snapshot.Up {
		t.Fatalf("snapshot must be up")
	}
	if snapshot.NodeTotal != 1 || snapshot.LeaseCount != 1 {
		t.Fatalf("unexpected snapshot counts: %+v", snapshot)
	}
	if len(snapshot.MetricNames) == 0 {
		t.Fatalf("metric names must not be empty")
	}
}

func TestNodeEndpoint(t *testing.T) {
	server := newTestServer(t)
	seedID := node.NodeID("seed-node")
	request := httptest.NewRequest(http.MethodGet, "/api/nodes/"+seedID, nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("node status = %d", recorder.Code)
	}
	missing := httptest.NewRequest(http.MethodGet, "/api/nodes/n-missing", nil)
	missingRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(missingRecorder, missing)
	if missingRecorder.Code != http.StatusNotFound {
		t.Fatalf("missing node status = %d", missingRecorder.Code)
	}
}
