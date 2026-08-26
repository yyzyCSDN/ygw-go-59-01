package model

import "time"

// NodeInfo 描述一个已注册的签发节点。
// Fingerprint 是节点名称的稳定散列，用于路由与对账。
type NodeInfo struct {
	ID           string
	Name         string
	Fingerprint  uint64
	RegisteredAt time.Time
	Active       bool
}

// NewNodeInfo 创建一条活跃的节点记录。
func NewNodeInfo(id, name string, fingerprint uint64, registeredAt time.Time) *NodeInfo {
	return &NodeInfo{
		ID:           id,
		Name:         name,
		Fingerprint:  fingerprint,
		RegisteredAt: registeredAt,
		Active:       true,
	}
}

// Deactivate 将节点标记为离线，禁止其继续持有签发权。
func (n *NodeInfo) Deactivate() {
	n.Active = false
}

// Reactivate 将节点重新标记为在线。
func (n *NodeInfo) Reactivate() {
	n.Active = true
}
