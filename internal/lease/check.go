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
	if lease.State != model.LeaseGranted {
		return model.ErrLeaseInvalid
	}
	if lease.Expired(m.now()) {
		return model.ErrLeaseInvalid
	}
	info, found := m.nodes.Lookup(lease.NodeID)
	if !found || !info.Active {
		return model.ErrNodeMissing
	}
	return nil
}

// ValidHolder 校验指定节点是否为该租约当前持有者。
// 用于签发链路二次确认，防止转移后的旧节点继续发号。
func (m *Manager) ValidHolder(leaseID, nodeID string) error {
	m.mu.Lock()
	lease, ok := m.leases[leaseID]
	m.mu.Unlock()
	if !ok || lease.State != model.LeaseGranted {
		return model.ErrLeaseInvalid
	}
	if lease.NodeID != nodeID {
		return model.ErrLeaseInvalid
	}
	return nil
}
