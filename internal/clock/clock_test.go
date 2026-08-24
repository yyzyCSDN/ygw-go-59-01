package clock

import (
	"testing"
	"time"

	"epochclock/internal/model"
)

func TestNewAndNow(t *testing.T) {
	hlc := New(42, 3)
	got := hlc.Now()
	if got.Physical != 42 || got.Logical != 3 {
		t.Fatalf("Now = %+v, want 42.3", got)
	}
}

func TestRestoreRaisesUpperBound(t *testing.T) {
	hlc := New(10, 0)
	restored := hlc.Restore(200, 5)
	if restored.Physical != 200 || restored.Logical != 5 {
		t.Fatalf("Restore = %+v, want 200.5", restored)
	}
	// 低于当前上限的恢复值不得拉低时钟。
	lower := hlc.Restore(5, 0)
	if lower.Physical != 200 {
		t.Fatalf("lower restore must not regress: %+v", lower)
	}
}

func TestWallMillis(t *testing.T) {
	instant := time.Unix(1700000000, 123000000)
	if got := WallMillis(instant); got != 1700000000123 {
		t.Fatalf("WallMillis = %d, want 1700000000123", got)
	}
	wall := SystemWall()
	if wall() == 0 {
		t.Fatalf("system wall must be non-zero")
	}
}

func TestClockPreservesModelType(t *testing.T) {
	var _ model.HLCTime = New(1, 1).Now()
}
