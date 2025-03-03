package kvstore

import (
	"raft"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]string
	raft *raft.RaftNode
}

func (s *Store) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.data[key]; ok {
		return val, nil
	}
	return "", ErrKeyNotFound
}

func (s *Store) Propose(key, value string) error {
	cmd := Command{
		Op:    "SET",
		Key:   key,
		Value: value,
	}
	return s.raft.Propose(cmd)
}

func (s *Store) ApplyLog(cmd Command) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch cmd.Op {
	case "SET":
		s.data[cmd.Key] = cmd.Value
	case "DELETE":
		delete(s.data, cmd.Key)
	}
}
