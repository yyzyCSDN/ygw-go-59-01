package persist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"epochclock/internal/model"
)

// Persister 是序列与时钟状态的持久化接口。
// NextRange 申请新区间并立即落盘；CommitRange 持久化区间游标；
// Replay 从 WAL 恢复区间、时钟与检查点。
type Persister interface {
	NextRange(start, size uint64) (*model.SeqRange, error)
	CommitRange(id, next uint64) error
	AppendClock(t model.HLCTime) error
	Replay() (*ReplayResult, error)
	Checkpoint(nextStart uint64) error
	SyncDir() error
	Close() error
	OpenSegments() int32
}

// WALPersister 基于追加式段文件实现持久化。
// 段文件达到 segmentSize 字节后轮转；已关闭的旧段可被清理，
// 未关闭的段会保留在 openFiles 中并被监控页统计。
type WALPersister struct {
	dir         string
	segmentSize int64
	mu          sync.Mutex
	openFiles   map[string]*os.File
	currentName string
	currentSeq  int64
	nextRangeID uint64
	closed      bool
	recordsWritten int64
	bytesWritten   int64
}

// NewWALPersister 创建持久化器并打开初始段文件。
func NewWALPersister(dir string, segmentSize int64) (*WALPersister, error) {
	if segmentSize <= 0 {
		segmentSize = 64 * 1024
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create wal dir: %w", err)
	}
	p := &WALPersister{
		dir:         dir,
		segmentSize: segmentSize,
		openFiles:   make(map[string]*os.File),
	}
	next, err := p.nextSegmentIndex()
	if err != nil {
		return nil, err
	}
	p.currentSeq = next
	if err := p.openSegmentLocked(); err != nil {
		return nil, err
	}
	return p, nil
}

// NextRange 分配并持久化一段新区间。
// 区间 ID 由持久化器维护，重启后从检查点与重放结果接续。
func (p *WALPersister) NextRange(start, size uint64) (*model.SeqRange, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, fmt.Errorf("%w: persister closed", model.ErrPersistence)
	}
	p.nextRangeID++
	if size == 0 {
		size = 100
	}
	end := start + size - 1
	r := model.NewSeqRange(p.nextRangeID, start, end)
	rec := model.RangeRecord(r, true)
	if err := p.appendLocked(rec); err != nil {
		return nil, err
	}
	return r, nil
}

// CommitRange 持久化区间游标推进结果，重放时以最新记录为准。
func (p *WALPersister) CommitRange(id, next uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return fmt.Errorf("%w: persister closed", model.ErrPersistence)
	}
	rec := model.CommitRecord(id, next)
	if err := p.appendLocked(rec); err != nil {
		return err
	}
	return nil
}

// AppendClock 持久化一次时钟推进。
func (p *WALPersister) AppendClock(t model.HLCTime) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return fmt.Errorf("%w: persister closed", model.ErrPersistence)
	}
	rec := model.ClockRecord(t)
	if err := p.appendLocked(rec); err != nil {
		return err
	}
	return nil
}

// Checkpoint 写入持久化检查点，重放时据此恢复下一个区间起始位置。
func (p *WALPersister) Checkpoint(nextStart uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return fmt.Errorf("%w: persister closed", model.ErrPersistence)
	}
	rec := model.WALRecord{Op: model.OpCheckpoint, Start: nextStart}
	if err := p.appendLocked(rec); err != nil {
		return err
	}
	return nil
}

// SyncDir 将数据目录元数据刷盘，保证新段文件名在崩溃后可见。
// Windows 上目录 fsync 不可用，返回 nil。
func (p *WALPersister) SyncDir() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	dir, err := os.Open(p.dir)
	if err != nil {
		return fmt.Errorf("%w: open dir: %v", model.ErrPersistence, err)
	}
	defer dir.Close()
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("%w: sync dir: %v", model.ErrPersistence, err)
	}
	return nil
}

// OpenSegments 返回当前打开中的段文件数量。
func (p *WALPersister) OpenSegments() int32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return int32(len(p.openFiles))
}

// Close 关闭全部段文件并释放资源。
func (p *WALPersister) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var firstErr error
	for name, f := range p.openFiles {
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(p.openFiles, name)
	}
	p.currentName = ""
	p.closed = true
	return firstErr
}

// encode 将记录序列化为一行 JSON。
func encode(rec model.WALRecord) ([]byte, error) {
	data, err := json.Marshal(rec)
	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", model.ErrPersistence, err)
	}
	return append(data, '\n'), nil
}

// segmentPath 返回段文件的绝对路径。
func (p *WALPersister) segmentPath(name string) string {
	return filepath.Join(p.dir, name)
}
