package verifycase

import (
	"errors"
	"testing"
	"time"

	"epochclock/internal/lease"
	"epochclock/internal/metric"
	"epochclock/internal/model"
	"epochclock/internal/node"
)

func TestMissingNodeNoNilPanic(t *testing.T) {
	metrics := metric.New()
	nodes := node.NewRegistry()
	info := nodes.Register("short-lived")
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	leases := lease.NewManager(nodes, func() time.Time { return now }, metrics)
	active, err := leases.Grant(info.ID, time.Minute)
	if err != nil {
		t.Fatalf("grant lease: %v", err)
	}
	// 节点随后从注册表移除，租约仍然存在。
	if !nodes.Unregister(info.ID) {
		t.Fatalf("unregister should succeed")
	}
	err = leases.CheckActive(active.ID)
	if !errors.Is(err, model.ErrNodeMissing) {
		t.Fatalf("lease check for a missing node must return ErrNodeMissing without panic, got %v", err)
	}
}
