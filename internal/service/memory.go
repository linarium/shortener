package service

import (
	"context"
	"sync"

	"github.com/linarium/shortener/internal/models"
)

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{data: make(map[string]string)}, nil
}

func (s *MemoryStorage) SaveShortURL(ctx context.Context, model models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[model.ShortURL] = model.OriginalURL
	return nil
}

func (s *MemoryStorage) GetLongURL(ctx context.Context, short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.data[short]
	return long, exists
}

func (s *MemoryStorage) Close() error {
	return nil
}
