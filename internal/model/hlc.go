package model

import "fmt"

// HLCTime 是混合逻辑时钟的一个时刻，由物理部分与逻辑部分组成。
// 比较时优先比较物理部分，物理部分相等再比较逻辑部分。
type HLCTime struct {
	Physical uint64
	Logical  uint32
}

// Before 判断当前时刻是否严格早于 other。
func (t HLCTime) Before(other HLCTime) bool {
	if t.Physical != other.Physical {
		return t.Physical < other.Physical
	}
	return t.Logical < other.Logical
}

// AtOrAfter 判断当前时刻是否不早于 other。
func (t HLCTime) AtOrAfter(other HLCTime) bool {
	return !t.Before(other)
}

// String 返回形如 physical.logical 的稳定文本表示。
func (t HLCTime) String() string {
	return fmt.Sprintf("%d.%d", t.Physical, t.Logical)
}

// Max 返回两个时刻中较晚的一个。
func Max(a, b HLCTime) HLCTime {
	if a.AtOrAfter(b) {
		return a
	}
	return b
}
