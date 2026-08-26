package alloc

import (
	"context"
	"fmt"

	"epochclock/internal/model"
	"epochclock/internal/persist"
)

// Issue 处理一次签发请求：校验取消与租约，分配序号并推进时钟。
// 返回的批次包含闭区间 [Start, End] 与签发时刻。
func (a *Allocator) Issue(ctx context.Context, leaseID, nodeID string, count uint64) (*model.IssuedBatch, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.leases.CheckActive(leaseID); err != nil {
		a.metrics.IncFailedIssue()
		return nil, err
	}
	if err := a.leases.ValidHolder(leaseID, nodeID); err != nil {
		a.metrics.IncFailedIssue()
		return nil, err
	}
	if count == 0 || count > a.batchSize {
		count = a.batchSize
	}
	if err := a.ensureRange(); err != nil {
		a.metrics.IncFailedIssue()
		return nil, err
	}
	batch, err := a.pool.BatchAlloc(count)
	if err != nil {
		a.metrics.IncRangeExhausted()
		a.metrics.IncFailedIssue()
		return nil, err
	}
	if batch.Start <= a.lastEnd {
		a.metrics.IncFailedIssue()
		return nil, fmt.Errorf("%w: start %d <= last issued end %d", model.ErrDuplicateRange, batch.Start, a.lastEnd)
	}
	if err := a.persist.CommitRange(batch.RangeID, a.pool.Current().Next); err != nil {
		a.metrics.IncFailedIssue()
		return nil, err
	}
	ts := a.nextTimestamp()
	batch.TS = ts
	a.lastTS = ts
	if err := a.persist.AppendClock(ts); err != nil {
		a.metrics.IncFailedIssue()
		return nil, err
	}
	a.lastEnd = batch.End
	a.metrics.IncIssued(batch.Count)
	a.RecordRecent(batch)
	return batch, nil
}

// checkCancelled 检查请求上下文是否已取消。
func (a *Allocator) checkCancelled(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: %v", model.ErrRequestCancelled, err)
	}
	return nil
}

// Restore 将 WAL 重放结果恢复进分配器：导入区间、接续起始序号、
// 恢复 HLC 物理上限与区间 ID。
func (a *Allocator) Restore(result *persist.ReplayResult) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, r := range result.Ranges {
		if err := a.pool.Add(r); err != nil {
			return err
		}
	}
	if result.NextStart > a.nextStart {
		a.nextStart = result.NextStart
	}
	if result.NextRangeID > a.nextRangeID {
		a.nextRangeID = result.NextRangeID
	}
	a.hlc.Restore(result.Clock.Physical, result.Clock.Logical)
	return nil
}

// SnapshotRanges 返回区间池快照，供状态接口使用。
func (a *Allocator) SnapshotRanges() []model.SeqRange {
	return a.pool.Snapshot()
}
