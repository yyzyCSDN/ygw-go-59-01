package verifycase

import (
	"testing"

	"epochclock/internal/clock"
)

func TestHLCMaxUpdatedOnAdvance(t *testing.T) {
	hlc := clock.New(0, 0)
	first := hlc.Advance(100)
	if first.Physical != 100 {
		t.Fatalf("advance must move the physical upper bound to the wall time, got %d", first.Physical)
	}
	// 物理时钟回拨：本地时刻 90 早于已观测最大值 100。
	second := hlc.Advance(90)
	if second.Physical != 100 {
		t.Fatalf("physical clock rollback must not regress the timestamp: got physical=%d", second.Physical)
	}
	if second.Logical <= first.Logical {
		t.Fatalf("logical part must keep advancing: first=%d second=%d", first.Logical, second.Logical)
	}
	got := hlc.Now()
	if got.Physical != 100 {
		t.Fatalf("clock upper bound was not retained: %+v", got)
	}
}
