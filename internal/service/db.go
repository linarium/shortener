package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/linarium/shortener/internal/models"
	"github.com/pressly/goose/v3"
)

type DB interface {
	Close() error
	PingContext(ctx context.Context) error
	sqlx.ExtContext
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
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

	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}

func applyMigrations(db DB) error {
	sqlDB, ok := db.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("expected *sqlx.DB, got %T", db)
	}

	if err := goose.Up(sqlDB.DB, "migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func (s *DBStorage) SaveShortURL(ctx context.Context, model models.URL) error {
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO urls (id, user_id, short_url, original_url)
        VALUES ($1, $2, $3, $4)
    `, model.ID, model.UserID, model.ShortURL, model.OriginalURL)
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
	var isDeleted bool

	err := s.db.QueryRowxContext(ctx, `
        SELECT original_url, is_deleted
        FROM urls
        WHERE short_url = $1
    `, short).Scan(&long, &isDeleted)

	if err != nil || isDeleted {
		return "", false
	}

	return long, true
}

func (s *DBStorage) SaveManyURLS(ctx context.Context, models []models.URL) error {
	query := `
        INSERT INTO urls (id, user_id, short_url, original_url)
        VALUES (:id, :user_id, :short_url, :original_url)
    `
	_, err := s.db.NamedExecContext(ctx, query, models)
	if err != nil {
		return fmt.Errorf("failed to save batch URLs: %w", err)
	}

	return nil
}

func (s *DBStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	var urls []models.URL

	query := `
		SELECT short_url, original_url 
		FROM urls 
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	err := s.db.SelectContext(ctx, &urls, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	return urls, nil
}

func (s *DBStorage) DeleteURLs(ctx context.Context, userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	placeholders := make([]string, len(shortURLs))
	args := make([]interface{}, len(shortURLs)+1)
	args[0] = userID

	for i, shortURL := range shortURLs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = shortURL
	}

	query := fmt.Sprintf(`
        UPDATE urls 
        SET is_deleted = TRUE 
        WHERE user_id = $1 
        AND short_url IN (%s)
    `, strings.Join(placeholders, ", "))

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete URLs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Soft deleted %d URLs for user %s\n", rowsAffected, userID)

	return nil
}
