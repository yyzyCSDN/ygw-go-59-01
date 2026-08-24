package metric

import (
	"testing"
)

func TestCounters(t *testing.T) {
	m := New()
	m.IncIssued(5)
	m.IncFailedIssue()
	m.IncRangeGranted()
	m.IncRangeExhausted()
	m.IncFailedRangeAlloc()
	m.IncClockAdvance()
	m.IncClockGuard()
	m.IncLeaseGrant()
	m.IncLeaseTransfer()
	m.IncLeaseExpired()
	m.AddReplayApplied(3)
	m.AddReplaySkipped(2)
	m.IncCancelled()

	snapshot := m.Snapshot()
	if snapshot["issued"] != 5 {
		t.Fatalf("issued = %d, want 5", snapshot["issued"])
	}
	if snapshot["replay_applied"] != 3 || snapshot["replay_skipped"] != 2 {
		t.Fatalf("replay counters wrong: %+v", snapshot)
	}
	expectedTotal := uint64(5 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 3 + 2 + 1)
	if m.Total() != expectedTotal {
		t.Fatalf("total = %d, want %d", m.Total(), expectedTotal)
	}
	names := Names(snapshot)
	if len(names) != len(snapshot) {
		t.Fatalf("names length mismatch")
	}
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Fatalf("names not sorted: %v", names)
		}
	}
	if got := Sum(snapshot, "failed_issues", "cancelled"); got != 2 {
		t.Fatalf("sum = %d, want 2", got)
	}
}
