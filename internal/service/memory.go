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

func (s *MemoryStorage) GetLongURL(ctx context.Context, short string) (string, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	long, exists := s.data[short]
	if !exists {
		return "", false, false
	}

	return long, true, false
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) SaveManyURLS(ctx context.Context, models []models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, model := range models {
		s.data[model.ShortURL] = model.OriginalURL
	}
	return nil
}

func (s *MemoryStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var urls []models.URL
	for shortURL, originalURL := range s.data {
		urls = append(urls, models.URL{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}

	return urls, nil
}

func (s *MemoryStorage) DeleteURLs(ctx context.Context, userID string, shortURLs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, shortURL := range shortURLs {
		delete(s.data, shortURL)
	}

	return nil
}
