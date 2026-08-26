package lease

import (
	"sort"

	"epochclock/internal/model"
)

// Snapshot 返回按租约 ID 排序的租约列表副本。
func (m *Manager) Snapshot() []model.Lease {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]model.Lease, 0, len(m.leases))
	for _, lease := range m.leases {
		items = append(items, *lease)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

// Count 返回当前管理的租约数量。
func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.leases)
}

// ActiveCount 返回当前处于 granted 状态的租约数量。
func (m *Manager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, lease := range m.leases {
		if lease.IsActive(m.now()) {
			count++
		}
	}
	return count
}
