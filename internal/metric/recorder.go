package metric

// Names 返回指标快照中的全部键名（稳定顺序）。
func Names(snapshot map[string]uint64) []string {
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

// Sum 返回快照中若干指定键的合计。
func Sum(snapshot map[string]uint64, keys ...string) uint64 {
	var total uint64
	for _, key := range keys {
		total += snapshot[key]
	}
	return total
}
