package node

import (
	"sort"
	"time"

	"epochclock/internal/model"
)

// Register 按名称注册一个新节点并返回其信息。
// 节点 ID 与指纹由名称稳定散列得到，重复注册返回已有节点。
func (r *Registry) Register(name string) *model.NodeInfo {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := NodeID(name)
	if existing, ok := r.nodes[id]; ok {
		return existing
	}
	info := model.NewNodeInfo(id, name, Fingerprint(name), time.Now().UTC())
	r.nodes[id] = info
	return info
}

// Snapshot 返回按节点 ID 排序的节点列表副本。
func (r *Registry) Snapshot() []model.NodeInfo {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]model.NodeInfo, 0, len(r.nodes))
	for _, info := range r.nodes {
		items = append(items, *info)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

// Deactivate 将节点标记为离线，使其不再满足签发前置条件。
func (r *Registry) Deactivate(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, ok := r.nodes[id]
	if !ok {
		return false
	}
	info.Deactivate()
	return true
}

// Reactivate 将节点重新标记为在线。
func (r *Registry) Reactivate(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, ok := r.nodes[id]
	if !ok {
		return false
	}
	info.Reactivate()
	return true
}

// Unregister 从注册表移除节点，常用于节点下线退役。
// 已移除节点的租约在签发前校验时会返回 ErrNodeMissing。
func (r *Registry) Unregister(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[id]; !ok {
		return false
	}
	delete(r.nodes, id)
	return true
}

// ActiveIDs 返回全部在线节点的 ID 列表。
func (r *Registry) ActiveIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ids []string
	for id, info := range r.nodes {
		if info.Active {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
