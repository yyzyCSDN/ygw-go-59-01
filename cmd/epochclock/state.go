package main

import (
	"time"

	"epochclock/internal/metric"
	"epochclock/internal/model"
	"epochclock/internal/persist"
)

// StatusSnapshot 是状态接口返回的整体运行快照。
type StatusSnapshot struct {
	Up           bool              `json:"up"`
	StartedAt    string            `json:"started_at"`
	Clock        model.HLCTime     `json:"clock"`
	Nodes        []model.NodeInfo  `json:"nodes"`
	Leases       []model.Lease     `json:"leases"`
	Ranges       []model.SeqRange  `json:"ranges"`
	Metrics      map[string]uint64 `json:"metrics"`
	MetricNames  []string          `json:"metric_names"`
	MetricTotal  uint64            `json:"metric_total"`
	RejectedTotal uint64           `json:"rejected_total"`
	OpenSegments int32             `json:"open_segments"`
	NextStart    uint64            `json:"next_start"`
	Remaining    uint64            `json:"remaining"`
	ActiveNodes  int               `json:"active_nodes"`
	NodeTotal    int               `json:"node_total"`
	LeaseCount   int               `json:"lease_count"`
	ActiveLeases int               `json:"active_leases"`
	Segments     []persist.SegmentInfo `json:"segments"`
	RecentBatches []model.IssuedBatch  `json:"recent_batches"`
	LeaseSummary  map[string]int       `json:"lease_summary"`
	Persist       persist.PersistSummary `json:"persist"`
	CurrentRange  *model.SeqRange      `json:"current_range"`
	TransferredLeases int              `json:"transferred_leases"`
}

// status 收集各组件当前状态。
func (s *Server) status() StatusSnapshot {
	metricsSnapshot := s.deps.Metrics.Snapshot()
	var segments []persist.SegmentInfo
	var persistSummary persist.PersistSummary
	if wal, ok := s.deps.Persister.(*persist.WALPersister); ok {
		segments = wal.SegmentSnapshot()
		persistSummary = wal.Summary()
	}
	return StatusSnapshot{
		Up:           true,
		StartedAt:    s.deps.StartedAt.Format(time.RFC3339),
		Clock:        s.deps.HLC.Now(),
		Nodes:        s.deps.Nodes.Snapshot(),
		Leases:       s.deps.Leases.Snapshot(),
		Ranges:       s.deps.Allocator.SnapshotRanges(),
		Metrics:      metricsSnapshot,
		MetricNames:  metric.Names(metricsSnapshot),
		MetricTotal:  s.deps.Metrics.Total(),
		RejectedTotal: metric.Sum(metricsSnapshot, "failed_issues", "cancelled"),
		OpenSegments: s.deps.Persister.OpenSegments(),
		NextStart:    s.deps.Allocator.NextStart(),
		Remaining:    s.deps.Allocator.Pool().Remaining(),
		ActiveNodes:  len(s.deps.Nodes.ActiveIDs()),
		NodeTotal:    s.deps.Nodes.Size(),
		LeaseCount:   s.deps.Leases.Count(),
		ActiveLeases: s.deps.Leases.ActiveCount(),
		Segments:     segments,
		RecentBatches: s.deps.Allocator.RecentBatches(),
		LeaseSummary:  s.deps.Leases.Summary(),
		Persist:       persistSummary,
		CurrentRange:  s.deps.Allocator.SequenceRange(),
		TransferredLeases: s.deps.Leases.TransferCount(),
	}
}
