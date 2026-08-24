package verifycase

import (
	"context"
	"testing"
	"time"

	"epochclock/internal/alloc"
	"epochclock/internal/clock"
	"epochclock/internal/lease"
	"epochclock/internal/metric"
	"epochclock/internal/model"
	"epochclock/internal/node"
	"epochclock/internal/persist"
)

type memPersister struct {
	nextRangeID uint64
}

func (m *memPersister) NextRange(start, size uint64) (*model.SeqRange, error) {
	m.nextRangeID++
	return model.NewSeqRange(m.nextRangeID, start, start+size-1), nil
}

func (m *memPersister) CommitRange(uint64, uint64) error { return nil }
func (m *memPersister) AppendClock(model.HLCTime) error  { return nil }
func (m *memPersister) Replay() (*persist.ReplayResult, error) {
	return &persist.ReplayResult{NextStart: 1}, nil
}
func (m *memPersister) Checkpoint(uint64) error { return nil }
func (m *memPersister) SyncDir() error          { return nil }
func (m *memPersister) Close() error            { return nil }
func (m *memPersister) OpenSegments() int32     { return 1 }

func TestBatchAllocNoDuplicate(t *testing.T) {
	metrics := metric.New()
	nodes := node.NewRegistry()
	info := nodes.Register("issuer")
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	leases := lease.NewManager(nodes, func() time.Time { return now }, metrics)
	active, err := leases.Grant(info.ID, time.Minute)
	if err != nil {
		t.Fatalf("grant lease: %v", err)
	}
	hlc := clock.New(100, 0)
	allocator := alloc.NewAllocator(&memPersister{nextRangeID: 1}, hlc, leases, metrics, func() uint64 { return 100 }, 10)
	replay := &persist.ReplayResult{Ranges: []*model.SeqRange{model.NewSeqRange(1, 1, 10)}, NextStart: 11, NextRangeID: 1}
	if err := allocator.Restore(replay); err != nil {
		t.Fatalf("restore: %v", err)
	}

	first, err := allocator.Issue(context.Background(), active.ID, info.ID, 3)
	if err != nil {
		t.Fatalf("first issue: %v", err)
	}
	second, err := allocator.Issue(context.Background(), active.ID, info.ID, 3)
	if err != nil {
		t.Fatalf("second issue: %v", err)
	}
	if second.Start != first.End+1 {
		t.Fatalf("batch boundary overlap: first=[%d,%d] second=[%d,%d]",
			first.Start, first.End, second.Start, second.End)
	}
	if second.Count != 3 {
		t.Fatalf("second batch must contain exactly 3 numbers, got %d", second.Count)
	}
}
