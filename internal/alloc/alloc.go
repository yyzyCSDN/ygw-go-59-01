package alloc

import (
	"fmt"
	"sync"

	"epochclock/internal/clock"
	"epochclock/internal/lease"
	"epochclock/internal/metric"
	"epochclock/internal/model"
	"epochclock/internal/persist"
)

// Allocator 是签发链路的核心：校验租约、申请区间、批量分配、
// 推进 HLC 并持久化游标。
type Allocator struct {
	mu          sync.Mutex
	pool        *RangePool
	persist     persist.Persister
	hlc         *clock.HLC
	leases      *lease.Manager
	metrics     *metric.Metrics
	wall        clock.WallClock
	batchSize   uint64
	nextStart   uint64
	nextRangeID uint64
	lastEnd     uint64
	lastTS      model.HLCTime
	recent      recentBatches
}

// NewAllocator 创建签发分配器。
func NewAllocator(
	pers persist.Persister,
	hlc *clock.HLC,
	leases *lease.Manager,
	metrics *metric.Metrics,
	wall clock.WallClock,
	batchSize uint64,
) *Allocator {
	return &Allocator{
		pool:      NewRangePool(),
		persist:   pers,
		hlc:       hlc,
		leases:    leases,
		metrics:   metrics,
		wall:      wall,
		batchSize: batchSize,
		nextStart: 1,
	}
}

// ensureRange 保证当前存在未用尽的区间；区间用尽时向持久化层
// 申请新区间并立即落盘，失败时向上返回错误。
func (a *Allocator) ensureRange() error {
	if r := a.pool.Current(); r != nil && !r.IsExhausted() {
		return nil
	}
	start := a.nextStart
	size := a.batchSize
	r, err := a.persist.NextRange(start, size)
	if err != nil {
		a.metrics.IncFailedRangeAlloc()
		return fmt.Errorf("%w: %v", model.ErrPersistence, err)
	}
	if err := a.pool.Add(r); err != nil {
		return err
	}
	if err := a.persist.SyncDir(); err != nil {
		return fmt.Errorf("%w: %v", model.ErrPersistence, err)
	}
	a.nextStart = r.End + 1
	a.nextRangeID = r.ID
	if err := a.persist.Checkpoint(a.nextStart); err != nil {
		return fmt.Errorf("%w: %v", model.ErrPersistence, err)
	}
	a.metrics.IncRangeGranted()
	return nil
}

// nextTimestamp 推进 HLC 并返回签发时刻。
// 返回值不得早于上一次签发时刻，否则触发单调保护并复用上一次时刻。
func (a *Allocator) nextTimestamp() model.HLCTime {
	now := a.wall()
	ts := a.hlc.Advance(now)
	if ts.Before(a.lastTS) {
		a.metrics.IncClockGuard()
		return a.lastTS
	}
	a.metrics.IncClockAdvance()
	return ts
}

// Pool 返回内部区间池，供监控快照使用。
func (a *Allocator) Pool() *RangePool {
	return a.pool
}

// NextStart 返回下一个新区间的起始序号。
func (a *Allocator) NextStart() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.nextStart
}
