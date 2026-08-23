/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : usage.json 聚合与最近请求持久化
 */
package usage

import (
	"fmt"
	"sync"
	"time"

	"codexrelay/internal/storage"
)

const (
	dataVersion       = 1
	maxRecentRequests = 300
	dayRetention      = 90
)

type Store struct {
	mu      sync.Mutex
	path    string
	data    Data
	persist bool
}

func NewStore(path string) (*Store, error) {
	store := &Store{path: path, persist: true}
	if err := store.loadOrCreate(); err != nil {
		return nil, err
	}
	return store, nil
}

// NewDeferredStore 创建首次引导期间使用的内存优先用量存储；读取已有文件，但不在引导完成前写回。
func NewDeferredStore(path string) (*Store, error) {
	store := &Store{path: path, persist: false}
	if err := store.loadOrCreate(); err != nil {
		return nil, err
	}
	return store, nil
}

// Path 返回当前用量文件路径；路径只用于运行时目录迁移和诊断，不对界面暴露文件内容。
func (s *Store) Path() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.path
}

// SetPath 在运行时目录迁移提交后切换后续用量写入目标；不会改变已有统计快照或持久化开关。
func (s *Store) SetPath(path string) {
	s.mu.Lock()
	s.path = path
	s.mu.Unlock()
}

func emptyData() Data {
	return Data{
		Version:   dataVersion,
		UpdatedAt: time.Now(),
		Profiles:  map[string]Aggregate{},
		Days:      map[string]Day{},
		Recent:    []RequestRecord{},
	}
}

func (s *Store) loadOrCreate() error {
	exists, err := storage.ReadJSON(s.path, &s.data)
	if err != nil {
		return err
	}
	if !exists {
		s.data = emptyData()
		if !s.persist {
			return nil
		}
		return s.saveLocked()
	}
	if s.data.Version != dataVersion {
		return fmt.Errorf("unsupported usage data version: %d", s.data.Version)
	}
	if s.data.Profiles == nil {
		s.data.Profiles = map[string]Aggregate{}
	}
	if s.data.Days == nil {
		s.data.Days = map[string]Day{}
	}
	if s.data.Recent == nil {
		s.data.Recent = []RequestRecord{}
	}
	if len(s.data.Recent) > maxRecentRequests {
		s.data.Recent = s.data.Recent[:maxRecentRequests]
	}
	s.trimDaysLocked(time.Now())
	return nil
}

func (s *Store) Record(record RequestRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Recent = append([]RequestRecord{record}, s.data.Recent...)
	if len(s.data.Recent) > maxRecentRequests {
		s.data.Recent = s.data.Recent[:maxRecentRequests]
	}
	addRecord(&s.data.Total, record)
	profile := s.data.Profiles[record.ProfileID]
	addRecord(&profile, record)
	s.data.Profiles[record.ProfileID] = profile

	dayKey := record.StartedAt.Local().Format("2006-01-02")
	day := s.data.Days[dayKey]
	if day.Profiles == nil {
		day.Profiles = map[string]Aggregate{}
	}
	addRecord(&day.Total, record)
	dayProfile := day.Profiles[record.ProfileID]
	addRecord(&dayProfile, record)
	day.Profiles[record.ProfileID] = dayProfile
	s.data.Days[dayKey] = day
	s.data.UpdatedAt = time.Now()
	s.trimDaysLocked(time.Now())
	return s.saveLocked()
}

func addRecord(aggregate *Aggregate, record RequestRecord) {
	aggregate.Requests++
	if record.UsageStatus != StatusReported {
		return
	}
	aggregate.ReportedRequests++
	aggregate.InputTokens += record.InputTokens
	aggregate.OutputTokens += record.OutputTokens
	aggregate.CachedTokens += record.CachedTokens
	aggregate.TotalTokens += record.TotalTokens
}

func (s *Store) Snapshot() Data {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneData(s.data)
}

func (s *Store) Overview() Overview {
	snapshot := s.Snapshot()
	return Overview{
		UpdatedAt: snapshot.UpdatedAt,
		Total:     snapshot.Total,
		Profiles:  snapshot.Profiles,
		Days:      snapshot.Days,
	}
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = emptyData()
	return s.saveLocked()
}

// ActivatePersistence 在首次引导完成后写入当前用量快照，并使后续记录恢复正常持久化。
func (s *Store) ActivatePersistence() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.persist {
		return nil
	}
	if err := s.writeLocked(); err != nil {
		return err
	}
	s.persist = true
	return nil
}

func (s *Store) trimDaysLocked(now time.Time) {
	cutoff := now.Local().AddDate(0, 0, -(dayRetention - 1)).Format("2006-01-02")
	for key := range s.data.Days {
		if key < cutoff {
			delete(s.data.Days, key)
		}
	}
}

func (s *Store) saveLocked() error {
	if !s.persist {
		return nil
	}
	return s.writeLocked()
}

func (s *Store) writeLocked() error {
	s.data.Version = dataVersion
	if err := storage.WriteJSONAtomic(s.path, ".usage-*.tmp", s.data); err != nil {
		return fmt.Errorf("保存用量数据: %w", err)
	}
	return nil
}

func cloneData(source Data) Data {
	clone := source
	clone.Profiles = make(map[string]Aggregate, len(source.Profiles))
	for id, aggregate := range source.Profiles {
		clone.Profiles[id] = aggregate
	}
	clone.Days = make(map[string]Day, len(source.Days))
	for date, day := range source.Days {
		dayClone := day
		dayClone.Profiles = make(map[string]Aggregate, len(day.Profiles))
		for id, aggregate := range day.Profiles {
			dayClone.Profiles[id] = aggregate
		}
		clone.Days[date] = dayClone
	}
	clone.Recent = append([]RequestRecord(nil), source.Recent...)
	return clone
}
