package persist

import "os"

// osStat 包装 os.Stat，便于在监控快照中复用。
func osStat(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
