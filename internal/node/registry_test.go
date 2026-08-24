package node

import (
	"testing"
)

func TestRegistryRegisterAndLookup(t *testing.T) {
	registry := NewRegistry()
	info := registry.Register("alpha")
	if info.ID != NodeID("alpha") {
		t.Fatalf("id mismatch: %s", info.ID)
	}
	again := registry.Register("alpha")
	if again.ID != info.ID {
		t.Fatalf("re-register must return the same node")
	}
	got, ok := registry.Lookup(info.ID)
	if !ok || got.Name != "alpha" {
		t.Fatalf("lookup failed: ok=%v got=%+v", ok, got)
	}
	if _, ok := registry.Lookup("n-missing"); ok {
		t.Fatalf("missing node must not be found")
	}
	if registry.Get("n-missing") != nil {
		t.Fatalf("missing node Get must return nil")
	}
}

func TestRegistrySnapshotAndState(t *testing.T) {
	registry := NewRegistry()
	registry.Register("alpha")
	registry.Register("beta")
	if registry.Size() != 2 {
		t.Fatalf("size = %d, want 2", registry.Size())
	}
	ids := registry.ActiveIDs()
	if len(ids) != 2 {
		t.Fatalf("active ids = %d, want 2", len(ids))
	}
	alphaID := NodeID("alpha")
	if !registry.Deactivate(alphaID) {
		t.Fatalf("deactivate should succeed")
	}
	if len(registry.ActiveIDs()) != 1 {
		t.Fatalf("deactivated node must leave active set")
	}
	if !registry.Reactivate(alphaID) {
		t.Fatalf("reactivate should succeed")
	}
	if len(registry.ActiveIDs()) != 2 {
		t.Fatalf("reactivated node must rejoin active set")
	}
	snapshot := registry.Snapshot()
	if len(snapshot) != 2 {
		t.Fatalf("snapshot length = %d, want 2", len(snapshot))
	}
}

func TestFingerprintStable(t *testing.T) {
	if Fingerprint("gamma") != Fingerprint("gamma") {
		t.Fatalf("fingerprint must be deterministic")
	}
	if Fingerprint("gamma") == Fingerprint("delta") {
		t.Fatalf("distinct names must have distinct fingerprints")
	}
	if len(NodeID("gamma")) == 0 {
		t.Fatalf("node id must not be empty")
	}
}

func TestUnregisterRemovesNode(t *testing.T) {
	registry := NewRegistry()
	info := registry.Register("retired")
	if !registry.Unregister(info.ID) {
		t.Fatalf("unregister should succeed for registered node")
	}
	if registry.Size() != 0 {
		t.Fatalf("size after unregister = %d, want 0", registry.Size())
	}
	if registry.Unregister(info.ID) {
		t.Fatalf("unregister should fail for missing node")
	}
}
