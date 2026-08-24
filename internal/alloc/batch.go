package alloc

import "epochclock/internal/model"

// BatchAlloc 从当前区间批量分配 count 个序号。
// 返回的闭区间 [Start, End] 恰好包含 count 个不重复序号，
// 区间游标推进到 End+1，下次分配从下一个序号开始。
func (p *RangePool) BatchAlloc(count uint64) (*model.IssuedBatch, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	r := p.current
	if r == nil || r.IsExhausted() {
		return nil, model.ErrRangeExhausted
	}
	from := r.Next
	if !r.Advance(count) {
		return nil, model.ErrRangeExhausted
	}
	to := from + count - 1
	return model.NewIssuedBatch(r.ID, from, to, model.HLCTime{}), nil
}

// Remaining 返回当前区间剩余可分配数量。
func (p *RangePool) Remaining() uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil {
		return 0
	}
	return p.current.Remaining()
}
