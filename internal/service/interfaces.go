package service

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

type Storage interface {
	SaveShortURL(short, long string)
	GetLongURL(short string) (string, bool)
}

type URLShortener struct {
	storage Storage
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
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
