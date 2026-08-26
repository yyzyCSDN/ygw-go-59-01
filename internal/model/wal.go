package model

// WALOp 标识 WAL 记录的操作类型。
type WALOp string

const (
	// OpRangeAllocated 记录新区间申请成功。
	OpRangeAllocated WALOp = "range_allocated"
	// OpRangeCommitted 记录区间游标推进并已提交。
	OpRangeCommitted WALOp = "range_committed"
	// OpClockAdvanced 记录 HLC 时钟推进。
	OpClockAdvanced WALOp = "clock_advanced"
	// OpCheckpoint 记录持久化检查点。
	OpCheckpoint WALOp = "checkpoint"
)

// WALRecord 是一条可持久化的 WAL 记录。
// Committed 为真表示该区间已提交，重放时不得重复应用。
type WALRecord struct {
	Op        WALOp  `json:"op"`
	RangeID   uint64 `json:"range_id,omitempty"`
	Start     uint64 `json:"start,omitempty"`
	End       uint64 `json:"end,omitempty"`
	Next      uint64 `json:"next,omitempty"`
	Committed bool   `json:"committed,omitempty"`
	Physical  uint64 `json:"physical,omitempty"`
	Logical   uint32 `json:"logical,omitempty"`
	LeaseID   string `json:"lease_id,omitempty"`
	NodeID    string `json:"node_id,omitempty"`
	Epoch     uint64 `json:"epoch,omitempty"`
}

// RangeRecord 构造一条区间分配记录。
func RangeRecord(r *SeqRange, committed bool) WALRecord {
	return WALRecord{
		Op:        OpRangeAllocated,
		RangeID:   r.ID,
		Start:     r.Start,
		End:       r.End,
		Next:      r.Next,
		Committed: committed,
	}
}

// CommitRecord 构造一条区间提交记录。
func CommitRecord(rangeID, next uint64) WALRecord {
	return WALRecord{
		Op:        OpRangeCommitted,
		RangeID:   rangeID,
		Next:      next,
		Committed: true,
	}
}

// ClockRecord 构造一条时钟推进记录。
func ClockRecord(t HLCTime) WALRecord {
	return WALRecord{
		Op:       OpClockAdvanced,
		Physical: t.Physical,
		Logical:  t.Logical,
	}
}
