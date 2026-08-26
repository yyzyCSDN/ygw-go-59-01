package node

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
)

// NodeID 由节点名称散列生成稳定的节点 ID。
func NodeID(name string) string {
	return fmt.Sprintf("n-%016x", Fingerprint(name))
}

// Fingerprint 返回节点名称的 xxhash 指纹，用于路由与对账。
func Fingerprint(name string) uint64 {
	return xxhash.Sum64String(name)
}
