package clock

import "epochclock/internal/model"

// Advance 用给定的本地物理时刻推进时钟并返回新时刻。
// 当本地时刻晚于已观测最大时刻时，物理上限前移、逻辑清零；
// 当本地时刻等于最大时刻时，仅逻辑递增；
// 当本地时刻早于最大时刻（物理回拨）时，保留物理上限并递增逻辑，
// 保证返回值仍然单调。
func (h *HLC) Advance(now uint64) model.HLCTime {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.advanceLocked(now)
}

// advanceLocked 是 Advance 的持锁实现。
func (h *HLC) advanceLocked(now uint64) model.HLCTime {
	if now > h.maxPhysical {
		h.logical++
		return model.HLCTime{Physical: h.maxPhysical, Logical: h.logical}
	}
	if now == h.maxPhysical {
		h.logical++
		return model.HLCTime{Physical: h.maxPhysical, Logical: h.logical}
	}
	// 物理时钟回拨：保留已观测到的物理上限，仅递增逻辑部分。
	h.logical++
	return model.HLCTime{Physical: h.maxPhysical, Logical: h.logical}
}
