package model

// IssuedBatch 是一次签发请求返回的序号区间与时间戳。
// Start 与 End 闭区间内所有序号本次一次性发放，不得重复。
type IssuedBatch struct {
	RangeID uint64
	Start   uint64
	End     uint64
	Count   uint64
	TS      HLCTime
}

// NewIssuedBatch 构造签发结果。
func NewIssuedBatch(rangeID, start, end uint64, ts HLCTime) *IssuedBatch {
	count := uint64(0)
	if end >= start {
		count = end - start + 1
	}
	return &IssuedBatch{
		RangeID: rangeID,
		Start:   start,
		End:     end,
		Count:   count,
		TS:      ts,
	}
}
