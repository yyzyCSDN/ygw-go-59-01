package lease

import (
	"fmt"
	"sync"
	"time"

	"epochclock/internal/metric"
	"epochclock/internal/model"
	"epochclock/internal/node"
)

// Manager 管理签发租约的授予、校验、过期与转移。
// 同一时刻一个租约最多被一个活跃节点持有。
type Manager struct {
	mu     sync.Mutex
	leases map[string]*model.Lease
	nodes  *node.Registry
	now    func() time.Time
	epoch  uint64
	metrics *metric.Metrics
}

// NewManager 创建租约管理器，now 可注入以便测试过期行为。
func NewManager(nodes *node.Registry, now func() time.Time, metrics *metric.Metrics) *Manager {
	return &Manager{
		leases: make(map[string]*model.Lease),
		nodes:  nodes,
		now:    now,
		metrics: metrics,
	}
}

// Grant 为已注册节点授予一条租约。
// 节点未注册时返回 ErrNodeMissing。
func (m *Manager) Grant(nodeID string, ttl time.Duration) (*model.Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.nodes.Lookup(nodeID); !ok {
		return nil, model.ErrNodeMissing
	}
	m.epoch++
	now := m.now()
	lease := model.NewLease(leaseID(m.epoch), nodeID, now, now.Add(ttl), m.epoch)
	lease.Activate()
	m.leases[lease.ID] = lease
	m.metrics.IncLeaseGrant()
	return lease, nil
}

// Expire 将指定租约标记为过期。
func (m *Manager) Expire(leaseID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	lease, ok := m.leases[leaseID]
	if !ok {
		return model.ErrLeaseInvalid
	}
	lease.ExpireNow()
	m.metrics.IncLeaseExpired()
	return nil
}

// leaseID 生成形如 l-<epoch> 的租约 ID。
func leaseID(epoch uint64) string {
	return fmt.Sprintf("l-%d", epoch)
}
