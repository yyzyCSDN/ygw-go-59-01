package metric

import "sync"

// Metrics 汇总签发、租约、持久化与时钟的运行计数。
// 所有指标在监控页与状态接口中展示。
type Metrics struct {
	mu               sync.Mutex
	issued           uint64
	failedIssues     uint64
	rangesGranted    uint64
	rangesExhausted  uint64
	failedRangeAlloc uint64
	clockAdvances    uint64
	clockGuards      uint64
	leaseGrants      uint64
	leaseTransfers   uint64
	leaseExpired     uint64
	replayApplied    uint64
	replaySkipped    uint64
	cancelled        uint64
}

// New 创建空的指标收集器。
func New() *Metrics {
	return &Metrics{}
}

// IncIssued 记录一次成功签发的序号数量。
func (m *Metrics) IncIssued(n uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.issued += n
}

// IncFailedIssue 记录一次失败的签发请求。
func (m *Metrics) IncFailedIssue() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedIssues++
}

// IncRangeGranted 记录一次新区间申请成功。
func (m *Metrics) IncRangeGranted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rangesGranted++
}

// IncRangeExhausted 记录一次区间耗尽事件。
func (m *Metrics) IncRangeExhausted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rangesExhausted++
}

// IncFailedRangeAlloc 记录一次区间申请失败。
func (m *Metrics) IncFailedRangeAlloc() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedRangeAlloc++
}

// IncClockAdvance 记录一次时钟正常推进。
func (m *Metrics) IncClockAdvance() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clockAdvances++
}

// IncClockGuard 记录一次时钟单调保护触发。
func (m *Metrics) IncClockGuard() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clockGuards++
}

// IncLeaseGrant 记录一次租约授予。
func (m *Metrics) IncLeaseGrant() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leaseGrants++
}

// IncLeaseTransfer 记录一次租约转移。
func (m *Metrics) IncLeaseTransfer() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leaseTransfers++
}

// IncLeaseExpired 记录一次租约到期。
func (m *Metrics) IncLeaseExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leaseExpired++
}

// AddReplayApplied 记录重放成功应用的区间数量。
func (m *Metrics) AddReplayApplied(n uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replayApplied += n
}

// AddReplaySkipped 记录重放去重跳过的记录数量。
func (m *Metrics) AddReplaySkipped(n uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replaySkipped += n
}

// IncCancelled 记录一次被取消的签发请求。
func (m *Metrics) IncCancelled() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelled++
}

// Snapshot 返回按名称排序的指标快照。
func (m *Metrics) Snapshot() map[string]uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]uint64{
		"issued":            m.issued,
		"failed_issues":     m.failedIssues,
		"ranges_granted":    m.rangesGranted,
		"ranges_exhausted":  m.rangesExhausted,
		"failed_range_alloc": m.failedRangeAlloc,
		"clock_advances":    m.clockAdvances,
		"clock_guards":      m.clockGuards,
		"lease_grants":      m.leaseGrants,
		"lease_transfers":   m.leaseTransfers,
		"lease_expired":     m.leaseExpired,
		"replay_applied":    m.replayApplied,
		"replay_skipped":    m.replaySkipped,
		"cancelled":         m.cancelled,
	}
}

// Total 返回全部指标数值之和，供监控页展示汇总值。
func (m *Metrics) Total() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.issued + m.failedIssues + m.rangesGranted + m.rangesExhausted +
		m.failedRangeAlloc + m.clockAdvances + m.clockGuards +
		m.leaseGrants + m.leaseTransfers + m.leaseExpired +
		m.replayApplied + m.replaySkipped + m.cancelled
}
