package lease

import "epochclock/internal/model"

// CheckActive 校验租约是否仍然有效：
// 租约必须存在、处于 granted 状态、未到期，且持有节点已注册并在线。
// 任一条件不满足都返回明确错误，签发链路据此停止分配。
func (m *Manager) CheckActive(leaseID string) error {
	m.mu.Lock()
	lease, ok := m.leases[leaseID]
	m.mu.Unlock()
	if !ok {
		return model.ErrLeaseInvalid
	}
	// 租约必须处于 granted 状态且尚未到期；否则签发会越过已观测的
	// lastEnd/lastTS 基线继续发号，产出回退的序号与时间戳，破坏下游单调排序。
	if !lease.IsActive(m.now()) {
		return model.ErrLeaseInvalid
	}
	info, found := m.nodes.Lookup(lease.NodeID)
	if !found || !info.Active {
		return model.ErrNodeMissing
	}
	return nil
}

// ValidHolder 校验指定节点是否为该租约当前持有者。
// 用于签发链路二次确认，防止转移或过期后的旧节点继续发号：
// 租约必须处于 granted 状态、未到期，且持有者与入参一致。
func (m *Manager) ValidHolder(leaseID, nodeID string) error {
	m.mu.Lock()
	lease, ok := m.leases[leaseID]
	m.mu.Unlock()
	if !ok || !lease.IsActive(m.now()) || lease.NodeID != nodeID {
		return model.ErrLeaseInvalid
	}
	return nil
}
