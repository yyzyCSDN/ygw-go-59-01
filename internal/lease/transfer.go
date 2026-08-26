package lease

import (
	"time"

	"epochclock/internal/model"
)

// Failover 是故障转移的完整流程：吊销旧节点持有的全部租约后，
// 为新节点签发一条新租约。返回新租约与吊销数量。
// 旧租约必须立即停止生效，否则新旧节点会同时签发导致序号冲突。
func (m *Manager) Failover(oldNodeID, newNodeID string, ttl time.Duration) (*model.Lease, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.nodes.Lookup(newNodeID); !ok {
		return nil, 0, model.ErrNodeMissing
	}
	now := m.now()
	m.epoch++
	next := model.NewLease(leaseID(m.epoch), newNodeID, now, now.Add(ttl), m.epoch)
	next.Activate()
	m.leases[next.ID] = next
	revoked := 0
	for id, lease := range m.leases {
		if lease.NodeID == oldNodeID && lease.State == model.LeaseGranted {
			lease.ExpireNow()
			m.leases[id].State = model.LeaseTransferred
			revoked++
		}
	}
	m.metrics.IncLeaseTransfer()
	return next, revoked, nil
}
