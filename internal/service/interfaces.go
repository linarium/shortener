package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/models"
)

func NewStorage(ctx context.Context, cfg config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return NewDBStorage(ctx, "pgx", cfg.DatabaseDSN)
	}

	if cfg.FileStoragePath != "" {
		return NewFileStorage(cfg.FileStoragePath)
	}

	return NewMemoryStorage(ctx)
}

type Storage interface {
	SaveShortURL(ctx context.Context, model models.URL) error
	GetLongURL(ctx context.Context, short string) (string, bool)
	Close() error
}

type Shortener interface {
	Shorten(ctx context.Context, url string) (string, error)
	Expand(ctx context.Context, shortURL string) (string, bool)
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

func (s *URLShortener) Shorten(ctx context.Context, longURL string) string {
	shortKey := s.generateShortKey()
	model := models.URL{
		ID:          uuid.New().String(),
		ShortURL:    shortKey,
		OriginalURL: longURL,
	}
	s.storage.SaveShortURL(ctx, model)
	return shortKey
}

func (s *URLShortener) Expand(ctx context.Context, shortKey string) (string, bool) {
	return s.storage.GetLongURL(ctx, shortKey)
}
