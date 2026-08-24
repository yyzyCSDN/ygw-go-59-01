package model

import "time"

// LeaseState 描述签发租约的生命周期。
type LeaseState string

const (
	// LeasePending 表示租约已创建但尚未生效。
	LeasePending LeaseState = "pending"
	// LeaseGranted 表示租约生效，持有节点可以签发序号。
	LeaseGranted LeaseState = "granted"
	// LeaseExpired 表示租约到期，持有节点停止签发。
	LeaseExpired LeaseState = "expired"
	// LeaseTransferred 表示租约已转移给其他节点，旧持有者停止签发。
	LeaseTransferred LeaseState = "transferred"
)

// Lease 是节点持有签发权的凭证。
// HolderEpoch 每次转移递增，用于区分新旧持有者。
type Lease struct {
	ID          string
	NodeID      string
	GrantedAt   time.Time
	ExpiresAt   time.Time
	State       LeaseState
	HolderEpoch uint64
}

// NewLease 创建处于 pending 状态的租约。
func NewLease(id, nodeID string, grantedAt, expiresAt time.Time, epoch uint64) *Lease {
	return &Lease{
		ID:          id,
		NodeID:      nodeID,
		GrantedAt:   grantedAt,
		ExpiresAt:   expiresAt,
		State:       LeasePending,
		HolderEpoch: epoch,
	}
}

// Expired 判断租约在给定时刻是否已经到期。
func (l *Lease) Expired(now time.Time) bool {
	return !now.Before(l.ExpiresAt)
}

// IsActive 判断租约当前是否处于可签发状态。
func (l *Lease) IsActive(now time.Time) bool {
	return l.State == LeaseGranted && !l.Expired(now)
}

// Activate 将 pending 租约置为 granted。
func (l *Lease) Activate() {
	if l.State == LeasePending {
		l.State = LeaseGranted
	}
}

// ExpireNow 将租约置为 expired。
func (l *Lease) ExpireNow() {
	l.State = LeaseExpired
}
