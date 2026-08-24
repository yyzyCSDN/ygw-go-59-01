package clock

import "epochclock/internal/model"

// Restore 从持久化恢复的检查点恢复时钟。
// 返回值表示恢复后的时钟时刻。
func (h *HLC) Restore(physical uint64, logical uint32) model.HLCTime {
	h.mu.Lock()
	defer h.mu.Unlock()
	if physical > h.maxPhysical {
		h.maxPhysical = physical
	}
	if physical == h.maxPhysical && logical > h.logical {
		h.logical = logical
	}
	return model.HLCTime{Physical: h.maxPhysical, Logical: h.logical}
}
