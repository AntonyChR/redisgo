package storage

import (
	"errors"
	"sync"
)

func NewStorage() *Storage {
	return &Storage{
		keyValueData:          make(map[string]string),
		keyListData:           make(map[string][]string),
		enabledRegisterOffset: false,
		registerOffset:        0,
		waiters: make(map[string][]chan string),
	}
}

type Storage struct {
	mu                    sync.RWMutex
	keyValueData          map[string]string
	keyListData           map[string][]string
	enabledRegisterOffset bool
	registerOffset        int
	waiters               map[string][]chan string
	waitersMux            sync.Mutex
}

func (s *Storage) RegisterWaiter(key string, ch chan string){
	s.waitersMux.Lock()
	defer s.waitersMux.Unlock()
	s.waiters[key] = append(s.waiters[key], ch)
}

func (s *Storage) NotifyWaiter(key, value string){
	s.waitersMux.Lock()
	defer s.waitersMux.Unlock()

	if waiters, ok := s.waiters[key];ok && len(waiters) > 0  {
		waiterChan := waiters[0]
		waiterChan <- value
		s.waiters[key] = waiters[1:]
	}
}

func (s *Storage) UnregisterWaiter(key string, ch chan string){
	s.waitersMux.Lock()
	defer s.waitersMux.Unlock()

	if waiters, ok := s.waiters[key]; ok {
		newWaiters := []chan string{}
		for _, waiter := range waiters {
			if waiter != ch {
				newWaiters = append(newWaiters, waiter)
			}
		}
		s.waiters[key] = newWaiters
	}
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

func (s *Storage) DeleteValue(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keyValueData, key)
}

func (s *Storage) AppendValuesToList(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keyListData[key] = append(s.keyListData[key], values...)
	return len(s.keyListData[key])
}

func (s *Storage) PrependValuesToList(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keyListData[key] = append(values, s.keyListData[key]...)
	return len(s.keyListData[key])
}

func (s *Storage) GetSliceFromList(key string, start, stop int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, ok := s.keyListData[key]
	if !ok {
		return []string{}
	}

	start, stop, err := nomralizeListIndexes(start, stop, len(list))
	if err != nil {
		return []string{}
	}

	return list[start : stop+1]
}

func (s *Storage) GetListLenght(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if list, ok := s.keyListData[key]; ok {
		return len(list)
	}
	return 0
}

func (s *Storage) RemoveElementFromListByIndex(key string, index int) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, ok := s.keyListData[key]
	if !ok {
		return ""
	}
	value := list[index]

	s.keyListData[key] =append(list[:index], list[index+1:]...)
	if len(s.keyListData[key]) == 0 {
		delete(s.keyListData, key)
	}
	return value
}

func (s *Storage) RemoveFirstElementFromTheList(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, ok := s.keyListData[key]
	if !ok {
		return ""
	}

	value := list[0]

	s.keyListData[key] = list[1:]
	if len(s.keyListData[key]) == 0 {
		delete(s.keyListData, key)
	}
	return value
}

func (s *Storage) RemoveFirstElementsFromTheList(key string, n int) []string {
	return s.RemoveElementsFromListByRange(key, 0, n)
}

func (s *Storage) RemoveElementsFromListByRange(key string, start, stop int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, ok := s.keyListData[key]
	if !ok {
		return []string{}
	}

	start, stop, err := nomralizeListIndexes(start, stop, len(list))
	if err != nil {
		return []string{}
	}

	removedElements := make([]string, stop-start+1)
	copy(removedElements, list[start:stop+1])
	s.keyListData[key] = append(list[:start], list[stop+1:]...)

	if len(s.keyListData[key]) == 0 {
		delete(s.keyListData, key)
	}
	
	return removedElements
}

func nomralizeListIndexes(start, stop, listLength int) (int, int, error) {
	if start < 0 {
		start += listLength
	}
	if stop < 0 {
		stop += listLength
	}

	start = max(0, start)
	stop = min(listLength-1, stop)

	if start > stop || start >= listLength {
		return 0, 0, errors.New("invalid range")
	}

	return start, stop, nil
}
