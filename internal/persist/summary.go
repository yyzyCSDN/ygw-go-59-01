package persist

// PersistSummary 是持久化层的运行摘要，供状态接口与监控页展示。
type PersistSummary struct {
	TotalSegments  int    `json:"total_segments"`
	OpenSegments   int32  `json:"open_segments"`
	RecordsWritten int64  `json:"records_written"`
	BytesWritten   int64  `json:"bytes_written"`
	NextRangeID    uint64 `json:"next_range_id"`
}

// Summary 返回持久化层运行摘要。
func (p *WALPersister) Summary() PersistSummary {
	p.mu.Lock()
	defer p.mu.Unlock()
	names, _ := p.segmentNames()
	return PersistSummary{
		TotalSegments:  len(names),
		OpenSegments:   int32(len(p.openFiles)),
		RecordsWritten: p.recordsWritten,
		BytesWritten:   p.bytesWritten,
		NextRangeID:    p.nextRangeID,
	}
}
