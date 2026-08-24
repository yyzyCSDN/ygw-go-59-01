package lease

import "epochclock/internal/model"

// Summary 返回按状态统计的租约数量。
func (m *Manager) Summary() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	summary := map[string]int{
		string(model.LeasePending):     0,
		string(model.LeaseGranted):     0,
		string(model.LeaseExpired):     0,
		string(model.LeaseTransferred): 0,
	}
	for _, lease := range m.leases {
		summary[string(lease.State)]++
	}
	return summary
}

// TransferCount 返回已转移租约的数量。
func (m *Manager) TransferCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, lease := range m.leases {
		if lease.State == model.LeaseTransferred {
			count++
		}
	}
	return count
}

// RevokeForNode 吊销指定节点持有的全部有效租约，节点注销时调用。
// 返回被吊销的租约数量。
func (m *Manager) RevokeForNode(nodeID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	revoked := 0
	for _, lease := range m.leases {
		if lease.NodeID == nodeID && lease.State == model.LeaseGranted {
			lease.ExpireNow()
			m.metrics.IncLeaseExpired()
			revoked++
		}
	}
	return revoked
}
