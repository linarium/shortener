package service

import "sync"

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (s *MemoryStorage) SaveShortURL(short, long string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[short] = long
}

func (s *MemoryStorage) GetLongURL(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.data[short]
	return long, exists
}
