package verifycase

import (
	"context"
	"errors"
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

func TestIssueHonorsCancel(t *testing.T) {
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
	allocator := alloc.NewAllocator(&memPersister{}, hlc, leases, metrics, func() uint64 { return 100 }, 10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = allocator.Issue(ctx, active.ID, info.ID, 3)
	if !errors.Is(err, model.ErrRequestCancelled) {
		t.Fatalf("cancelled issue must not allocate a sequence, got %v", err)
	}
}
