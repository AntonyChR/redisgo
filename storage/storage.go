package storage

import "errors"



func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
		enabledRegisterOffset: false,
		registerOffset: 0,
	}
}

type Storage struct{
	data map[string]string
	enabledRegisterOffset bool
	registerOffset int
}

func (s *Storage) Get(key string) (value string, err error){
	value, ok := s.data[key]

	if !ok {
		return "", errors.New("value not found")
	}

	return value, nil 
}


func (s *Storage) Set(key, value string) {
	s.data[key] = value
}


func (s *Storage) Delete(key string){
	delete(s.data, key)
}

