package node

import (
	"sync"

	"epochclock/internal/model"
)

// Registry 维护已注册的签发节点。
// 节点注册后即可申请租约；未注册节点发起签发时应返回明确错误。
type Registry struct {
	mu    sync.Mutex
	nodes map[string]*model.NodeInfo
}

// NewRegistry 创建空的节点注册表。
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]*model.NodeInfo),
	}
}

// Get 返回指定节点信息。节点不存在时返回 nil。
func (r *Registry) Get(id string) *model.NodeInfo {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.nodes[id]
}

// Lookup 返回指定节点信息及其存在性。
// 调用方应优先使用 Lookup，避免对缺失节点解引用。
func (r *Registry) Lookup(id string) (*model.NodeInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, ok := r.nodes[id]
	return info, ok
}

// Size 返回已注册节点数量。
func (r *Registry) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.nodes)
}
