package storage

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	utils "github.com/AntonyChR/go-utils"
)

func NewStorage() *Storage {
	return &Storage{
		keyValueData:          make(map[string]string),
		keyListData:           make(map[string][]string),
		streamData: 		   make(map[string][]map[string]string),
		enabledRegisterOffset: false,
		registerOffset:        0,
		waiters: 			   make(map[string][]chan string),
	}
}


type Storage struct {
	mu                    sync.RWMutex

	//data
	keyValueData          map[string]string
	keyListData           map[string][]string
	streamData            map[string][]map[string]string

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

// streamData methods
func (s *Storage) GetLastEntryStream(key string) (entry map[string]string, listLen int){
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.streamData[key]
	if !ok {
		return nil, 0
	}
	entry = l[len(l) - 1]
	return entry, len(l)
}

func (s *Storage) AddEntryStream(key string, data map[string]string) error{
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.streamData[key]; 
	if !ok {
		s.streamData[key] = []map[string]string{data}
		return nil
	}

	s.streamData[key] = append(s.streamData[key], data)	
	return nil
}

func (s *Storage) GetStreamEntriesByRange(key string,startTimestamp, endTimestamp int64, startIndex, endIndex int) []map[string]string{
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	list, ok := s.streamData[key]
	if !ok {
		return []map[string]string{}
	}

	cb := func(s map[string]string)bool{
		timestamp, index := parseEntryId(s["id"])				
		if startIndex== 0 && endIndex == 0 {
			timeStampIsInRange := startTimestamp <= timestamp && timestamp <= endTimestamp
			return timeStampIsInRange 
		}
		indexIsInRange := startIndex <= index && index <= endIndex
		timeStampIsInRange := startTimestamp <= timestamp && timestamp <= endTimestamp
		return timeStampIsInRange && indexIsInRange
	}

	if endTimestamp == 0 && endIndex == 0 {
		cb =  func(s map[string]string)bool{
			timestamp, index:= parseEntryId(s["id"])				

			timeStampIsInRange := startTimestamp <= timestamp && timestamp <= endTimestamp
			indexIsInRange := startIndex <= index 
			return timeStampIsInRange && indexIsInRange
		}
	}
	
	filteredData := utils.FilterPrealloc(list, cb)
	resp := utils.DeepCopyArrMap(filteredData)
	return resp
}


func (s *Storage) GetStreamEntriesByPartialRange(key string,startTimestamp int64, startIndex int) []map[string]string{
	s.mu.RLock()
	defer s.mu.RUnlock()
	list, ok := s.streamData[key]
	if !ok {
		return []map[string]string{}
	}

	cb := func(s map[string]string)bool{
		timestamp, index := parseEntryId(s["id"])				
		return startTimestamp < timestamp ||  startIndex < index 
	}

	filteredData := utils.FilterPrealloc(list, cb)
	resp := utils.DeepCopyArrMap(filteredData)
	return resp
}


func parseEntryId(id string) (int64, int){
	s := strings.Split(id, "-")
	timeStamp,_ := strconv.ParseInt(s[0], 10,64) 
	index ,_ := strconv.Atoi(s[1]) 
	return timeStamp, index
}

func (s *Storage) CheckType(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if _, ok := s.keyValueData[key]; ok {
		return "string"
	}

	if _, ok := s.streamData[key]; ok {
		return "stream"
	}

	return "none"
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
