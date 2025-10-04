package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/models"
	"github.com/linarium/shortener/internal/service"
	"strings"
)

type Repository interface {
	Shorten(ctx context.Context, url string, userID string) (string, bool)
	ShortenBatch(ctx context.Context, longs models.BatchRequest, baseURL string, userID string) (models.BatchResponse, error)
	Expand(ctx context.Context, shortURL string) (string, bool)
	Ping(ctx context.Context) error
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)
}

type ShortenerService struct {
	storage service.Storage
}

func NewShortenerService(storage service.Storage) Repository {
	return &ShortenerService{storage: storage}
}

func (s *ShortenerService) generateShortKey() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		logger.Sugar.Fatalf("Ошибка при генерации сокращённого URL: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)[:8]
}

func (s *ShortenerService) Shorten(ctx context.Context, longURL string, userID string) (string, bool) {
	shortKey := s.generateShortKey()
	model := models.URL{
		ID:          uuid.New().String(),
		ShortURL:    shortKey,
		OriginalURL: longURL,
		UserID:      userID,
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

func (s *ShortenerService) Expand(ctx context.Context, shortKey string) (string, bool) {
	return s.storage.GetLongURL(ctx, shortKey)
}

func (s *ShortenerService) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *ShortenerService) findShortKeyByOriginalURL(ctx context.Context, original string) (string, bool) {
	return s.storage.FindShortURLByOriginal(ctx, original)
}

func (s *ShortenerService) ShortenBatch(ctx context.Context, longs models.BatchRequest, baseURL string, userID string) (models.BatchResponse, error) {
	length := len(longs)
	shorts := make(models.BatchResponse, length)
	urls := make([]models.URL, length)

	for i, long := range longs {
		shortKey := s.generateShortKey()
		urls[i] = models.URL{
			ID:          uuid.New().String(),
			ShortURL:    shortKey,
			OriginalURL: long.OriginalURL,
			UserID:      userID,
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

func (s *ShortenerService) GetUserURLs(ctx context.Context, userID string) ([]models.URL, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	return s.storage.GetAll(ctx, userID)
}
