package verifycase

import (
	"testing"

	"epochclock/internal/persist"
)

func TestReplayNoDuplicateApply(t *testing.T) {
	dir := t.TempDir()
	wal, err := persist.NewWALPersister(dir, 1024)
	if err != nil {
		t.Fatalf("open wal: %v", err)
	}
	defer func() {
		_ = wal.Close()
	}()

	first, err := wal.NextRange(1, 100)
	if err != nil {
		t.Fatalf("NextRange: %v", err)
	}
	if err := wal.CommitRange(first.ID, 10); err != nil {
		t.Fatalf("CommitRange: %v", err)
	}

	result, err := wal.Replay()
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(result.Ranges) != 1 {
		t.Fatalf("replay must apply a committed range exactly once, got %d ranges", len(result.Ranges))
	}
	if result.Ranges[0].Next != 10 {
		t.Fatalf("replayed range must restore the committed cursor, got Next=%d", result.Ranges[0].Next)
	}
	if result.Applied != 1 {
		t.Fatalf("applied count must be 1, got %d", result.Applied)
	}
}
