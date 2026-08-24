package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHLCTimeOrdering(t *testing.T) {
	early := HLCTime{Physical: 100, Logical: 1}
	late := HLCTime{Physical: 100, Logical: 2}
	later := HLCTime{Physical: 101, Logical: 0}
	if !early.Before(late) {
		t.Fatalf("expected %s before %s", early, late)
	}
	if !late.Before(later) {
		t.Fatalf("expected %s before %s", late, later)
	}
	if !later.AtOrAfter(early) {
		t.Fatalf("expected %s at-or-after %s", later, early)
	}
	if got := Max(early, later); got != later {
		t.Fatalf("Max mismatch: got %s want %s", got, later)
	}
	if got := early.String(); got != "100.1" {
		t.Fatalf("String mismatch: %q", got)
	}
}

func TestSeqRangeLifecycle(t *testing.T) {
	r := NewSeqRange(7, 100, 105)
	if r.Remaining() != 6 {
		t.Fatalf("remaining = %d, want 6", r.Remaining())
	}
	if r.IsExhausted() {
		t.Fatalf("fresh range must not be exhausted")
	}
	if !r.Advance(2) {
		t.Fatalf("advance 2 should succeed")
	}
	if r.Next != 102 || r.State != RangeDraining {
		t.Fatalf("after advance: next=%d state=%s", r.Next, r.State)
	}
	if r.Advance(4) {
		t.Fatalf("advance to range end must report exhaustion")
	}
	if !r.IsExhausted() || r.State != RangeExhausted {
		t.Fatalf("range should be exhausted: state=%s", r.State)
	}
	if r.Remaining() != 0 {
		t.Fatalf("remaining must be 0, got %d", r.Remaining())
	}
	if r.Advance(1) {
		t.Fatalf("advance on exhausted range must fail")
	}
}

func TestLeaseStateTransitions(t *testing.T) {
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	lease := NewLease("l-1", "n-1", now, now.Add(time.Minute), 1)
	if lease.State != LeasePending {
		t.Fatalf("new lease must be pending")
	}
	lease.Activate()
	if !lease.IsActive(now.Add(time.Second)) {
		t.Fatalf("activated lease must be active")
	}
	if lease.IsActive(now.Add(2 * time.Minute)) {
		t.Fatalf("expired lease must not be active")
	}
	lease.ExpireNow()
	if lease.State != LeaseExpired {
		t.Fatalf("expected expired state")
	}
}

func TestWALRecordRoundTrip(t *testing.T) {
	original := WALRecord{
		Op:        OpRangeCommitted,
		RangeID:   3,
		Start:     10,
		End:       19,
		Next:      12,
		Committed: true,
		Physical:  1700000000000,
		Logical:   4,
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded WALRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded != original {
		t.Fatalf("round trip mismatch: %+v != %+v", decoded, original)
	}
}

func TestRecordBuilders(t *testing.T) {
	r := NewSeqRange(2, 50, 59)
	allocated := RangeRecord(r, true)
	if allocated.Op != OpRangeAllocated || !allocated.Committed || allocated.Start != 50 {
		t.Fatalf("bad allocated record: %+v", allocated)
	}
	committed := CommitRecord(2, 55)
	if committed.Op != OpRangeCommitted || committed.Next != 55 || !committed.Committed {
		t.Fatalf("bad committed record: %+v", committed)
	}
	clockRecord := ClockRecord(HLCTime{Physical: 99, Logical: 1})
	if clockRecord.Op != OpClockAdvanced || clockRecord.Physical != 99 {
		t.Fatalf("bad clock record: %+v", clockRecord)
	}
}

func TestSentinelErrors(t *testing.T) {
	if ErrRangeExhausted == nil || ErrLeaseInvalid == nil || ErrNodeMissing == nil ||
		ErrPersistence == nil || ErrRequestCancelled == nil || ErrDuplicateRange == nil {
		t.Fatalf("sentinel errors must be non-nil")
	}
}

func TestIssuedBatch(t *testing.T) {
	batch := NewIssuedBatch(1, 5, 8, HLCTime{Physical: 1, Logical: 0})
	if batch.Count != 4 || batch.Start != 5 || batch.End != 8 {
		t.Fatalf("bad batch: %+v", batch)
	}
}
