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

func (s *Storage) SetValueToList(key string, values ...string) int{
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keyListData[key] = append(s.keyListData[key], values...)
	return len(s.keyListData[key])
}

func (s *Storage) GetSliceFromList(key string, start, end int) []string{
	values := make([]string, 0)

	if list, ok := s.keyListData[key]; ok {

		if start >= len(list) - 1 {
			return values
		}

		if start > end {
			return values
		}

		if end >= len(list) - 1 {
			end = len(list)
		}  
		
		values = list[start:end]
	}

	return values
}
func (s *Storage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keyValueData, key)
}
