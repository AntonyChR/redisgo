package storage

import "sync"

func NewStorage() *Storage {
	return &Storage{
		data:                  make(map[string]string),
		enabledRegisterOffset: false,
		registerOffset:        0,
	}
}

type Storage struct {
	mu                    sync.RWMutex
	data                  map[string]string
	enabledRegisterOffset bool
	registerOffset        int
}

func (s *Storage) Get(key string) (value string, exists bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]
	if !ok {
		return "", false
	}
	return value, true

}

func (s *Storage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Storage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}
