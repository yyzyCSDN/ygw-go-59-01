package persist

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"epochclock/internal/model"
)

var segmentPattern = regexp.MustCompile(`^wal-(\d{6})\.seg$`)

// appendLocked 向当前段写入一条记录，必要时轮转段文件。
func (p *WALPersister) appendLocked(rec model.WALRecord) error {
	if p.currentName == "" {
		if err := p.openSegmentLocked(); err != nil {
			return err
		}
	}
	data, err := encode(rec)
	if err != nil {
		return err
	}
	f := p.openFiles[p.currentName]
	if f == nil {
		return fmt.Errorf("%w: current segment not open", model.ErrPersistence)
	}
	written, err := f.Write(data)
	if err != nil {
		return fmt.Errorf("%w: write: %v", model.ErrPersistence, err)
	}
	p.recordsWritten++
	p.bytesWritten += int64(written)
	if err := f.Sync(); err != nil {
		return fmt.Errorf("%w: sync: %v", model.ErrPersistence, err)
	}
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("%w: stat: %v", model.ErrPersistence, err)
	}
	if info.Size() >= p.segmentSize {
		return p.rotateLocked()
	}
	return nil
}

// rotateLocked 关闭当前段并打开下一段。
// 旧段关闭后可从 openFiles 移除并可被清理。
func (p *WALPersister) rotateLocked() error {
	if p.currentName != "" {
		if f, ok := p.openFiles[p.currentName]; ok {
			if err := f.Close(); err != nil {
				return fmt.Errorf("%w: close segment: %v", model.ErrPersistence, err)
			}
			delete(p.openFiles, p.currentName)
		}
		p.currentName = ""
	}
	p.pruneSegmentsLocked()
	return p.openSegmentLocked()
}

// openSegmentLocked 打开一个新的段文件并登记到 openFiles。
func (p *WALPersister) openSegmentLocked() error {
	p.currentSeq++
	name := fmt.Sprintf("wal-%06d.seg", p.currentSeq)
	f, err := os.OpenFile(p.segmentPath(name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("%w: open segment: %v", model.ErrPersistence, err)
	}
	p.openFiles[name] = f
	p.currentName = name
	return nil
}

// nextSegmentIndex 返回现有段文件中的最大序号。
func (p *WALPersister) nextSegmentIndex() (int64, error) {
	names, err := p.segmentNames()
	if err != nil {
		return 0, err
	}
	max := int64(0)
	for _, name := range names {
		matches := segmentPattern.FindStringSubmatch(name)
		if len(matches) != 2 {
			continue
		}
		index, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			continue
		}
		if index > max {
			max = index
		}
	}
	return max, nil
}

// segmentNames 返回目录中按名称排序的段文件列表。
func (p *WALPersister) segmentNames() ([]string, error) {
	entries, err := os.ReadDir(p.dir)
	if err != nil {
		return nil, fmt.Errorf("%w: list dir: %v", model.ErrPersistence, err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !segmentPattern.MatchString(entry.Name()) {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

// pruneSegmentsLocked 删除旧的已关闭段文件，保留当前段与最近两个段。
// 仍在 openFiles 中的段不删除，避免并发读写时文件被占用。
func (p *WALPersister) pruneSegmentsLocked() {
	names, err := p.segmentNames()
	if err != nil {
		return
	}
	if len(names) <= 3 {
		return
	}
	for _, name := range names[:len(names)-3] {
		if _, open := p.openFiles[name]; open {
			continue
		}
		_ = os.Remove(filepath.Join(p.dir, name))
	}
}
