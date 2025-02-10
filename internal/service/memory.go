package service

import (
	"github.com/linarium/shortener/internal/models"
	"sync"
)

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{data: make(map[string]string)}, nil
}

func (s *MemoryStorage) SaveShortURL(model models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[model.ShortURL] = model.OriginalURL
	return nil
}

func (s *MemoryStorage) GetLongURL(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.data[short]
	return long, exists
}

func (s *MemoryStorage) Close() error {
	return nil
}
