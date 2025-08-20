package storage

import (
	"sync"
)

func NewStorage() *Storage {
	return &Storage{
		keyValueData:          make(map[string]string),
		keyListData:           make(map[string][]string),
		enabledRegisterOffset: false,
		registerOffset:        0,
	}
}

type Storage struct {
	mu                    sync.RWMutex
	keyValueData          map[string]string
	keyListData           map[string][]string
	enabledRegisterOffset bool
	registerOffset        int
}

func (s *Storage) Get(key string) (value string, exists bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.keyValueData[key]
	if !ok {
		return "", false
	}
	return value, true

}

func (s *Storage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keyValueData[key] = value
}

func (s *Storage) SetValueToList(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keyListData[key] = append(s.keyListData[key], values...)
	return len(s.keyListData[key])
}

func (s *Storage) GetSliceFromList(key string, start, stop int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, ok := s.keyListData[key]
	if !ok {
		return []string{}
	}

	if start < 0 {
		start += len(list)
	}
	if stop < 0 {
		stop += len(list)
	}

	start = max(0, start)
	stop = min(len(list)-1, stop)

	if start > stop || start >= len(list) {
		return []string{}
	}

	return list[start : stop+1]
}

func (s *Storage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keyValueData, key)
}
