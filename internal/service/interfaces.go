package service

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/models"
)

type Shortener interface {
	Shorten(url string) (string, error)
	Expand(shortURL string) (string, bool)
}

type URLShortener struct {
	storage Storage
}

func NewURLShortener(storage Storage) *URLShortener {
	return &URLShortener{storage: storage}
}

func (s *URLShortener) generateShortKey() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		logger.Sugar.Fatalf("Ошибка при генерации сокращённого URL: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)[:8]
}

func (s *URLShortener) Shorten(longURL string) string {
	shortKey := s.generateShortKey()
	model := models.URL{ShortURL: shortKey, OriginalURL: longURL}
	s.storage.SaveShortURL(model)
	return shortKey
}

func (s *URLShortener) Expand(shortKey string) (string, bool) {
	return s.storage.GetLongURL(shortKey)
}
