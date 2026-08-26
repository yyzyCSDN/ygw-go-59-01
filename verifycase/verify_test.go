package verifycase

import (
	"testing"

	"epochclock/internal/persist"
)

func TestWALHandleClosed(t *testing.T) {
	dir := t.TempDir()
	wal, err := persist.NewWALPersister(dir, 180)
	if err != nil {
		t.Fatalf("open wal: %v", err)
	}
	defer func() {
		_ = wal.Close()
	}()

	// 持续写入并轮转多个段文件。
	for i := 0; i < 10; i++ {
		if _, err := wal.NextRange(uint64(1+i*100), 100); err != nil {
			t.Fatalf("NextRange %d: %v", i, err)
		}
	}
	open := wal.OpenSegments()
	if open > 1 {
		t.Fatalf("old WAL segment handles must be closed after rotation, open segments = %d", open)
	}
}
