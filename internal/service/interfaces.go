package service

import (
	"context"
	"github.com/linarium/shortener/internal/config"
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
	GetAll(ctx context.Context, userID string) ([]models.URL, error)
	GetLongURL(ctx context.Context, short string) (string, bool, bool)
	FindShortURLByOriginal(ctx context.Context, original string) (string, bool)
	Ping(ctx context.Context) error
	Close() error
	DeleteURLs(ctx context.Context, userID string, shortURLs []string) error
}

func (s *DBStorage) FindShortURLByOriginal(ctx context.Context, original string) (string, bool) {
	var short string
	err := s.db.QueryRowxContext(ctx, `SELECT short_url FROM urls WHERE original_url = $1 LIMIT 1`, original).Scan(&short)
	if err != nil {
		return "", false
	}
	return short, true
}

func (s *MemoryStorage) FindShortURLByOriginal(ctx context.Context, original string) (string, bool) {
	for k, v := range s.data {
		if v == original {
			return k, true
		}
	}
	return "", false
}

func (s *FileStorage) FindShortURLByOriginal(ctx context.Context, original string) (string, bool) {
	return s.memory.FindShortURLByOriginal(ctx, original)
}
