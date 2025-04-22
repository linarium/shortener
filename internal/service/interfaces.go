package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/models"
)

func NewStorage(ctx context.Context, cfg config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return NewDBStorage(ctx, cfg.DatabaseDSN)
	}

	if cfg.FileStoragePath != "" {
		return NewFileStorage(cfg.FileStoragePath)
	}

	return NewMemoryStorage(ctx)
}

type Storage interface {
	SaveShortURL(ctx context.Context, model models.URL) error
	SaveManyURLS(ctx context.Context, models []models.URL) error
	GetLongURL(ctx context.Context, short string) (string, bool)
	Close() error
}

type Shortener interface {
	Shorten(ctx context.Context, url string) (string, error)
	ShortenBatch(ctx context.Context, longs models.BatchRequest, baseURL string) (models.BatchResponse, error)
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

func (s *URLShortener) Shorten(ctx context.Context, longURL string) (string, bool) {
	shortKey := s.generateShortKey()
	model := models.URL{
		ID:          uuid.New().String(),
		ShortURL:    shortKey,
		OriginalURL: longURL,
	}

	err := s.storage.SaveShortURL(ctx, model)
	if err != nil {
		if strings.HasPrefix(err.Error(), "duplicate_original:") {
			if existingShort, ok := s.findShortKeyByOriginalURL(ctx, longURL); ok {
				return existingShort, true
			}
		}
	}

	return shortKey, false
}

func (s *URLShortener) findShortKeyByOriginalURL(ctx context.Context, original string) (string, bool) {
	if dbStorage, ok := s.storage.(*DBStorage); ok {
		var short string
		err := dbStorage.db.QueryRowxContext(ctx, `
            SELECT short_url FROM urls WHERE original_url = $1 LIMIT 1
        `, original).Scan(&short)

		if err != nil {
			return "", false
		}
		return short, true
	}

	return "", false
}

func (s *URLShortener) Expand(ctx context.Context, shortKey string) (string, bool) {
	return s.storage.GetLongURL(ctx, shortKey)
}

func (s *URLShortener) ShortenBatch(ctx context.Context, longs models.BatchRequest, baseURL string) (models.BatchResponse, error) {
	length := len(longs)
	shorts := make(models.BatchResponse, length)
	urls := make([]models.URL, length)

	for i, long := range longs {
		shortKey := s.generateShortKey()
		urls[i] = models.URL{
			ID:          uuid.New().String(),
			ShortURL:    shortKey,
			OriginalURL: long.OriginalURL,
		}
		shorts[i] = models.BatchResponseItem{
			CorrelationID: long.CorrelationID,
			ShortURL:      baseURL + "/" + shortKey,
		}
	}

	if batchStorage, ok := s.storage.(interface {
		SaveManyURLS(ctx context.Context, urls []models.URL) error
	}); ok {
		err := batchStorage.SaveManyURLS(ctx, urls)
		if err != nil {
			return nil, fmt.Errorf("failed to save batch: %w", err)
		}
	} else {
		for _, url := range urls {
			if err := s.storage.SaveShortURL(ctx, url); err != nil {
				return nil, fmt.Errorf("failed to save URL: %w", err)
			}
		}
	}

	return shorts, nil
}
