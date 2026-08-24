package persist

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"epochclock/internal/model"
)

// ReplayResult 是 WAL 重放后恢复的运行时状态。
type ReplayResult struct {
	Ranges      []*model.SeqRange
	Clock       model.HLCTime
	NextStart   uint64
	NextRangeID uint64
	Applied     int
	Skipped     int
}

// Replay 读取全部段文件并恢复状态。
// 同一区间的多条记录只应用最新一条，已提交区间不会被重复应用。
func (p *WALPersister) Replay() (*ReplayResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := &ReplayResult{NextStart: 1}
	names, err := p.segmentNames()
	if err != nil {
		return nil, err
	}
	latest := make(map[uint64]model.WALRecord)
	var maxClock model.HLCTime
	for _, name := range names {
		records, err := p.readSegmentLocked(name)
		if err != nil {
			return nil, err
		}
		for _, rec := range records {
			switch rec.Op {
			case model.OpRangeAllocated, model.OpRangeCommitted:
				if rec.Committed {
					latest[rec.RangeID] = rec
				}
			case model.OpClockAdvanced:
				clock := model.HLCTime{Physical: rec.Physical, Logical: rec.Logical}
				maxClock = model.Max(maxClock, clock)
			case model.OpCheckpoint:
				if rec.Start > result.NextStart {
					result.NextStart = rec.Start
				}
			}
		}
	}
	ids := make([]uint64, 0, len(latest))
	for id := range latest {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		rec := latest[id]
		r := &model.SeqRange{
			ID:    rec.RangeID,
			Start: rec.Start,
			End:   rec.End,
			Next:  rec.Next,
			State: model.RangeOpen,
		}
		if rec.Next > rec.End {
			r.State = model.RangeExhausted
		} else if rec.Next > rec.Start {
			r.State = model.RangeDraining
		}
		result.Ranges = append(result.Ranges, r)
		if rec.RangeID > result.NextRangeID {
			result.NextRangeID = rec.RangeID
		}
		result.Applied++
	}
	result.Clock = maxClock
	if result.NextStart == 0 {
		result.NextStart = 1
	}
	return result, nil
}

// readSegmentLocked 读取单个段文件的全部记录。
func (p *WALPersister) readSegmentLocked(name string) ([]model.WALRecord, error) {
	f, err := os.Open(p.segmentPath(name))
	if err != nil {
		return nil, fmt.Errorf("%w: open segment %s: %v", model.ErrPersistence, name, err)
	}
	defer f.Close()
	var records []model.WALRecord
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec model.WALRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("%w: decode %s: %v", model.ErrPersistence, name, err)
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: read %s: %v", model.ErrPersistence, name, err)
	}
	return records, nil
}
