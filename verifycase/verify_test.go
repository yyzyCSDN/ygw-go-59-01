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

func TestFailoverStopsOldIssuer(t *testing.T) {
	metrics := metric.New()
	nodes := node.NewRegistry()
	oldNode := nodes.Register("old-issuer")
	newNode := nodes.Register("new-issuer")
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	leases := lease.NewManager(nodes, func() time.Time { return now }, metrics)
	oldLease, err := leases.Grant(oldNode.ID, time.Minute)
	if err != nil {
		t.Fatalf("grant old lease: %v", err)
	}

	newLease, revoked, err := leases.Failover(oldNode.ID, newNode.ID, time.Minute)
	if err != nil {
		t.Fatalf("failover: %v", err)
	}
	if newLease.NodeID != newNode.ID {
		t.Fatalf("new lease must belong to the new node, got %s", newLease.NodeID)
	}
	if revoked == 0 {
		t.Fatalf("failover must revoke the old issuer's lease")
	}
	err = leases.CheckActive(oldLease.ID)
	if !errors.Is(err, model.ErrLeaseInvalid) {
		t.Fatalf("old lease must stop issuing after failover, got %v", err)
	}
}
