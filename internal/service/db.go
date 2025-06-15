package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/linarium/shortener/internal/models"
)

type DB interface {
	Close() error
	PingContext(ctx context.Context) error
	sqlx.ExtContext
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

func NewDB(DatabaseDSName string) (DB, error) {
	return sqlx.Open("pgx", DatabaseDSName)
}

type DBStorage struct {
	db DB
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func NewDBStorage(ctx context.Context, dataSourceName string) (*DBStorage, error) {
	db, err := NewDB(dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}

func (s *DBStorage) SaveShortURL(ctx context.Context, model models.URL) error {
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO urls (id, short_url, original_url)
        VALUES ($1, $2, $3)
    `, model.ID, model.ShortURL, model.OriginalURL)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("duplicate_original:%s", model.OriginalURL)
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}

	return nil
}

func (s *DBStorage) GetLongURL(ctx context.Context, short string) (string, bool) {
	var long string
	err := s.db.QueryRowxContext(ctx, `
        SELECT original_url
        FROM urls
        WHERE short_url = $1
    `, short).Scan(&long)

	if err != nil {
		return "", false
	}

	return long, true
}

func (s *DBStorage) SaveManyURLS(ctx context.Context, models []models.URL) error {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO urls (id, short_url, original_url)
        VALUES (:id, :short_url, :original_url)
	`, models)
	if err != nil {
		return err
	}

	return nil
}
