package service

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"sync"
)

type URLStorage interface {
	SaveShortURL(short, long string)
	GetLongURL(short string) (string, bool)
}

type MemoryStorage struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{urls: make(map[string]string)}
}

func (s *MemoryStorage) SaveShortURL(short, long string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[short] = long
}

func (s *MemoryStorage) GetLongURL(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.urls[short]
	return long, exists
}

type URLShortener struct {
	storage URLStorage
}

func NewURLShortener() *URLShortener {
	return &URLShortener{storage: NewMemoryStorage()}
}

func (s *URLShortener) generateShortKey() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Ошибка при генерации сокращённого URL: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)[:8]
}

func (s *URLShortener) Shorten(longURL string) string {
	shortKey := s.generateShortKey()
	s.storage.SaveShortURL(shortKey, longURL)
	return shortKey
}

func (s *URLShortener) Expand(shortKey string) (string, bool) {
	return s.storage.GetLongURL(shortKey)
}
