package raft

import (
	"compress/gzip"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	snapshotDir      = "snapshots"
	snapshotInterval = 10 * time.Minute
	retainSnapshots  = 3
)

type Snapshot struct {
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

type SnapshotMetadata struct {
	Index int
	Term  int
	Size  int64
	Time  time.Time
}

type SnapshotStore struct {
	mu        sync.RWMutex
	latest    *SnapshotMetadata
	snapshots []SnapshotMetadata
	compress  bool
	raftNode  *RaftNode
}

func NewSnapshotStore(raftNode *RaftNode, compress bool) *SnapshotStore {
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		panic(err)
	}

	store := &SnapshotStore{
		compress: compress,
		raftNode: raftNode,
	}

	store.loadLatestSnapshot()
	go store.periodicSnapshot()
	return store
}

func (s *SnapshotStore) CreateSnapshot(index int, term int, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index <= s.latest.Index {
		return errors.New("snapshot index <= latest")
	}

	snapshot := Snapshot{
		LastIncludedIndex: index,
		LastIncludedTerm:  term,
		Data:              data,
	}

	filename := filepath.Join(snapshotDir,
		time.Now().UTC().Format("20060102T150405Z.snap"))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var writer io.Writer = file
	if s.compress {
		gz := gzip.NewWriter(file)
		defer gz.Close()
		writer = gz
	}

	enc := gob.NewEncoder(writer)
	if err := enc.Encode(snapshot); err != nil {
		os.Remove(filename)
		return err
	}

	// Обновление метаданных
	info, _ := file.Stat()
	metadata := SnapshotMetadata{
		Index: index,
		Term:  term,
		Size:  info.Size(),
		Time:  time.Now(),
	}

	s.snapshots = append(s.snapshots, metadata)
	s.latest = &metadata

	// Очистка старых снапшотов
	s.cleanupOldSnapshots()
	return nil
}

func (s *SnapshotStore) loadLatestSnapshot() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := filepath.Glob(filepath.Join(snapshotDir, "*.snap"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	// Загрузка последнего снапшота
	latestFile := files[len(files)-1]
	file, err := os.Open(latestFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var reader io.Reader = file
	if s.compress {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gz.Close()
		reader = gz
	}

	var snapshot Snapshot
	dec := gob.NewDecoder(reader)
	if err := dec.Decode(&snapshot); err != nil {
		return err
	}

	info, _ := file.Stat()
	s.latest = &SnapshotMetadata{
		Index: snapshot.LastIncludedIndex,
		Term:  snapshot.LastIncludedTerm,
		Size:  info.Size(),
		Time:  info.ModTime(),
	}

	// Восстановление состояния
	s.raftNode.restoreFromSnapshot(snapshot)
	return nil
}

func (s *SnapshotStore) GetSnapshot(index int) (*Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, meta := range s.snapshots {
		if meta.Index == index {
			filename := filepath.Join(snapshotDir,
				meta.Time.Format("20060102T150405Z.snap"))

			file, err := os.Open(filename)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			var reader io.Reader = file
			if s.compress {
				gz, err := gzip.NewReader(file)
				if err != nil {
					return nil, err
				}
				defer gz.Close()
				reader = gz
			}

			var snapshot Snapshot
			dec := gob.NewDecoder(reader)
			if err := dec.Decode(&snapshot); err != nil {
				return nil, err
			}

			return &snapshot, nil
		}
	}

	return nil, errors.New("snapshot not found")
}

func (s *SnapshotStore) cleanupOldSnapshots() {
	if len(s.snapshots) <= retainSnapshots {
		return
	}

	toDelete := len(s.snapshots) - retainSnapshots
	for _, meta := range s.snapshots[:toDelete] {
		filename := filepath.Join(snapshotDir,
			meta.Time.Format("20060102T150405Z.snap"))
		os.Remove(filename)
	}

	s.snapshots = s.snapshots[toDelete:]
}

func (s *SnapshotStore) periodicSnapshot() {
	ticker := time.NewTicker(snapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.raftNode.mu.Lock()
			if s.raftNode.state == Leader {
				s.raftNode.takeSnapshot()
			}
			s.raftNode.mu.Unlock()
		case <-s.raftNode.stopCh:
			return
		}
	}
}
