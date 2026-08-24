package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"epochclock/internal/model"
	"epochclock/web"
)

// issueRequest 是签发接口的请求体。
type issueRequest struct {
	NodeID  string `json:"node_id"`
	LeaseID string `json:"lease_id"`
	Count   uint64 `json:"count"`
}

// issueResponse 是签发接口的响应体。
type issueResponse struct {
	RangeID uint64        `json:"range_id"`
	Start   uint64        `json:"start"`
	End     uint64        `json:"end"`
	Count   uint64        `json:"count"`
	TS      model.HLCTime `json:"ts"`
}

// failoverRequest 是故障转移接口的请求体。
type failoverRequest struct {
	OldNodeID string `json:"old_node_id"`
	NewNodeID string `json:"new_node_id"`
}

// failoverResponse 是故障转移接口的响应体。
type failoverResponse struct {
	NewLeaseID string `json:"new_lease_id"`
	NodeID     string `json:"node_id"`
	Revoked    int    `json:"revoked"`
}

// expireRequest 是租约过期接口的请求体。
type expireRequest struct {
	LeaseID string `json:"lease_id"`
}

// unregisterRequest 是节点注销接口的请求体。
type unregisterRequest struct {
	NodeID string `json:"node_id"`
}

// handleHealth 返回健康检查结果。
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleStatus 返回运行状态快照。
func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.status()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(snapshot)
}

// handleNode 返回单个节点信息；节点不存在时返回 404。
func (s *Server) handleNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	info := s.deps.Nodes.Get(id)
	if info == nil {
		writeError(w, http.StatusNotFound, model.ErrNodeMissing.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(info)
}

// handleRanges 返回全部序列区间快照。
func (s *Server) handleRanges(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ranges":     s.deps.Allocator.SnapshotRanges(),
		"remaining":  s.deps.Allocator.Pool().Remaining(),
		"next_start": s.deps.Allocator.NextStart(),
	})
}

// handleLeases 返回全部租约快照与状态汇总。
func (s *Server) handleLeases(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"leases":  s.deps.Leases.Snapshot(),
		"summary": s.deps.Leases.Summary(),
	})
}

// handleNodeUnregister 注销节点，使其不再参与签发。
func (s *Server) handleNodeUnregister(w http.ResponseWriter, r *http.Request) {
	var req unregisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NodeID == "" {
		writeError(w, http.StatusBadRequest, "node_id is required")
		return
	}
	if !s.deps.Nodes.Unregister(req.NodeID) {
		writeError(w, http.StatusNotFound, model.ErrNodeMissing.Error())
		return
	}
	revoked := s.deps.Leases.RevokeForNode(req.NodeID)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"node_id": req.NodeID,
		"revoked_leases": revoked,
	})
}

// handleIssue 处理序列号签发请求。
func (s *Server) handleIssue(w http.ResponseWriter, r *http.Request) {
	var req issueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	batch, err := s.deps.Allocator.Issue(ctx, req.LeaseID, req.NodeID, req.Count)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, model.ErrLeaseInvalid):
			status = http.StatusForbidden
		case errors.Is(err, model.ErrNodeMissing):
			status = http.StatusNotFound
		case errors.Is(err, model.ErrRequestCancelled):
			status = http.StatusRequestTimeout
		case errors.Is(err, model.ErrRangeExhausted):
			status = http.StatusServiceUnavailable
		case errors.Is(err, model.ErrDuplicateRange):
			status = http.StatusConflict
		}
		log.Printf("issue failed: %v", err)
		writeError(w, status, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(issueResponse{
		RangeID: batch.RangeID,
		Start:   batch.Start,
		End:     batch.End,
		Count:   batch.Count,
		TS:      batch.TS,
	})
}

// handleMonitor 返回浏览器监控页面。
func (s *Server) handleMonitor(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(web.MonitorHTML())
}

// handleFailover 处理签发节点故障转移。
// 旧节点被吊销全部租约并标记离线，新节点获得新租约。
func (s *Server) handleFailover(w http.ResponseWriter, r *http.Request) {
	var req failoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OldNodeID == "" || req.NewNodeID == "" {
		writeError(w, http.StatusBadRequest, "old_node_id and new_node_id are required")
		return
	}
	deactivated := s.deps.Nodes.Deactivate(req.OldNodeID)
	newLease, revoked, err := s.deps.Leases.Failover(req.OldNodeID, req.NewNodeID, 30*time.Second)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.deps.Nodes.Reactivate(req.NewNodeID)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(failoverResponse{
		NewLeaseID: newLease.ID,
		NodeID:     newLease.NodeID,
		Revoked:    revoked,
	})
	log.Printf("failover complete: old=%s deactivated=%v new=%s lease=%s revoked=%d",
		req.OldNodeID, deactivated, req.NewNodeID, newLease.ID, revoked)
}

// handleLeaseExpire 手动将指定租约标记为过期。
func (s *Server) handleLeaseExpire(w http.ResponseWriter, r *http.Request) {
	var req expireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LeaseID == "" {
		writeError(w, http.StatusBadRequest, "lease_id is required")
		return
	}
	if err := s.deps.Leases.Expire(req.LeaseID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"lease_id": req.LeaseID, "state": "expired"})
}

// handleIndex 将根路径重定向到监控页面。
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/monitor", http.StatusFound)
}

// writeError 输出统一格式的错误响应。
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
