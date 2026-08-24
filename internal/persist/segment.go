package persist

// SegmentInfo 描述一个 WAL 段文件的运行状态，供监控页展示。
type SegmentInfo struct {
	Name      string `json:"name"`
	Open      bool   `json:"open"`
	Index     int64  `json:"index"`
	Bytes     int64  `json:"bytes"`
}

// SegmentSnapshot 返回全部段文件的状态列表。
func (p *WALPersister) SegmentSnapshot() []SegmentInfo {
	p.mu.Lock()
	defer p.mu.Unlock()
	names, err := p.segmentNames()
	if err != nil {
		return nil
	}
	items := make([]SegmentInfo, 0, len(names))
	for _, name := range names {
		index := int64(0)
		if matches := segmentPattern.FindStringSubmatch(name); len(matches) == 2 {
			index = parseIndex(matches[1])
		}
		_, open := p.openFiles[name]
		size := int64(0)
		if info, err := p.fileSize(name); err == nil {
			size = info
		}
		items = append(items, SegmentInfo{
			Name:  name,
			Open:  open,
			Index: index,
			Bytes: size,
		})
	}
	return items
}

// fileSize 返回段文件字节数。
func (p *WALPersister) fileSize(name string) (int64, error) {
	info, err := osStat(p.segmentPath(name))
	return info, err
}

// parseIndex 将段文件序号文本解析为整数。
func parseIndex(text string) int64 {
	var value int64
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			break
		}
		value = value*10 + int64(ch-'0')
	}
	return value
}
