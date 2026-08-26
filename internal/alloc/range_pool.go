package alloc

import (
	"sort"
	"sync"

	"epochclock/internal/model"
)

// RangePool 持有当前可分配区间与历史区间列表。
type RangePool struct {
	mu      sync.Mutex
	current *model.SeqRange
	ranges  []*model.SeqRange
}

// NewRangePool 创建空的区间池。
func NewRangePool() *RangePool {
	return &RangePool{}
}

// Current 返回当前可分配区间，没有区间时返回 nil。
func (p *RangePool) Current() *model.SeqRange {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

// Add 追加一个区间并设为当前区间；区间 ID 重复时拒绝。
func (p *RangePool) Add(r *model.SeqRange) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, existing := range p.ranges {
		if existing.ID == r.ID {
			return model.ErrDuplicateRange
		}
	}
	p.ranges = append(p.ranges, r)
	p.current = r
	return nil
}

// Snapshot 返回按区间 ID 排序的区间列表副本。
func (p *RangePool) Snapshot() []model.SeqRange {
	p.mu.Lock()
	defer p.mu.Unlock()
	items := make([]model.SeqRange, 0, len(p.ranges))
	for _, r := range p.ranges {
		items = append(items, *r)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}
