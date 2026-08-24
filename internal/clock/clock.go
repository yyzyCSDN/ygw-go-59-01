package clock

import (
	"sync"
	"time"

	"epochclock/internal/model"
)

// WallClock 返回当前物理时刻（毫秒）。
// 可注入以便在测试中精确控制物理时钟回拨与前进。
type WallClock func() uint64

// HLC 是混合逻辑时钟。maxPhysical 记录观察到的最大物理时刻，
// logical 记录同一物理时刻内的逻辑序号。物理时钟回拨时，
// maxPhysical 不回退，从而保证签发的时刻单调。
type HLC struct {
	mu          sync.Mutex
	maxPhysical uint64
	logical     uint32
}

// New 以指定物理与逻辑初值创建时钟。
func New(physical uint64, logical uint32) *HLC {
	return &HLC{
		maxPhysical: physical,
		logical:     logical,
	}
}

// Now 返回当前时钟时刻，不改变时钟状态。
func (h *HLC) Now() model.HLCTime {
	h.mu.Lock()
	defer h.mu.Unlock()
	return model.HLCTime{Physical: h.maxPhysical, Logical: h.logical}
}

// SystemWall 返回使用系统时钟的物理时刻来源。
func SystemWall() WallClock {
	return func() uint64 {
		return uint64(time.Now().UnixNano()) / 1e6
	}
}

// WallMillis 将 time.Time 转换为毫秒物理时刻。
func WallMillis(t time.Time) uint64 {
	return uint64(t.UnixNano()) / 1e6
}
