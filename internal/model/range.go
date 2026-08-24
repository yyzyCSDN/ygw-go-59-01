package model

// RangeState 描述序列区间生命周期。
type RangeState string

const (
	// RangeOpen 表示区间刚申请成功，游标在起始位置。
	RangeOpen RangeState = "open"
	// RangeDraining 表示区间已被部分消费，仍可继续分配。
	RangeDraining RangeState = "draining"
	// RangeExhausted 表示区间已用尽，不能再分配。
	RangeExhausted RangeState = "exhausted"
)

// SeqRange 是一段连续的序列号区间。
// Next 指向下一个待分配序号，取值在 [Start, End+1] 之间。
type SeqRange struct {
	ID    uint64
	Start uint64
	End   uint64
	Next  uint64
	State RangeState
}

// NewSeqRange 创建一段从 start 到 end 的开放区间。
func NewSeqRange(id, start, end uint64) *SeqRange {
	return &SeqRange{
		ID:    id,
		Start: start,
		End:   end,
		Next:  start,
		State: RangeOpen,
	}
}

// Remaining 返回区间内尚未分配的序号数量。
func (r *SeqRange) Remaining() uint64 {
	if r.Next > r.End {
		return 0
	}
	return r.End - r.Next + 1
}

// IsExhausted 判断区间是否已经用尽。
func (r *SeqRange) IsExhausted() bool {
	return r.Next > r.End
}

// MarkDraining 将开放区间标记为排空中的区间。
func (r *SeqRange) MarkDraining() {
	if r.State == RangeOpen {
		r.State = RangeDraining
	}
}

// MarkExhausted 将区间标记为已用尽。
func (r *SeqRange) MarkExhausted() {
	r.State = RangeExhausted
}

// Advance 按给定数量推进区间游标，返回是否仍然可用。
func (r *SeqRange) Advance(count uint64) bool {
	if r.Remaining() < count {
		return false
	}
	r.Next += count
	if r.IsExhausted() {
		r.MarkExhausted()
		return false
	}
	r.MarkDraining()
	return true
}
