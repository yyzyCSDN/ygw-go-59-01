package alloc

import (
	"sync"

	"epochclock/internal/model"
)

const maxRecentBatches = 16

// recentBatches 保存最近成功签发的批次，供监控页查看签发连续性。
type recentBatches struct {
	mu     sync.Mutex
	items  []model.IssuedBatch
	next   int
	filled bool
}

// add 记录一个新批次；超过上限时覆盖最旧的记录。
func (r *recentBatches) add(batch model.IssuedBatch) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) < maxRecentBatches {
		r.items = append(r.items, batch)
		return
	}
	r.items[r.next] = batch
	r.next = (r.next + 1) % maxRecentBatches
	r.filled = true
}

// snapshot 返回按签发顺序排列的批次副本。
func (r *recentBatches) snapshot() []model.IssuedBatch {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]model.IssuedBatch, 0, len(r.items))
	if r.filled {
		out = append(out, r.items[r.next:]...)
		out = append(out, r.items[:r.next]...)
	} else {
		out = append(out, r.items...)
	}
	return out
}

// RecentBatches 返回最近成功签发的批次列表。
func (a *Allocator) RecentBatches() []model.IssuedBatch {
	return a.recent.snapshot()
}

// RecordRecent 将一次成功签发写入近批记录。
func (a *Allocator) RecordRecent(batch *model.IssuedBatch) {
	if batch == nil {
		return
	}
	a.recent.add(*batch)
}

// SequenceRange 返回当前区间的运行摘要。
func (a *Allocator) SequenceRange() *model.SeqRange {
	return a.pool.Current()
}
