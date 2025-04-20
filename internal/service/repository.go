package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/linarium/shortener/internal/models"
)

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	memory MemoryStorage
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		model := &models.URL{}
		if err := json.Unmarshal(scanner.Bytes(), model); err != nil {
			return nil, err
		}
		data[model.ShortURL] = model.OriginalURL
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &FileStorage{
		file:   file,
		writer: bufio.NewWriter(file),
		memory: MemoryStorage{data: data},
	}, nil
}

func (s *FileStorage) GetLongURL(ctx context.Context, short string) (string, bool) {
	return s.memory.GetLongURL(ctx, short)
}

func (s *FileStorage) SaveShortURL(ctx context.Context, model models.URL) error {
	if err := json.NewEncoder(s.writer).Encode(model); err != nil {
		return err
	}
	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.SaveShortURL(ctx, model)
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}

type DBStorage struct {
	db DB
}

func NewDBStorage(ctx context.Context, driverName, dataSourceName string) (*DBStorage, error) {
	db, err := NewDB(dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = initDB(ctx, db); err != nil {
		return nil, err
	}

	return &DBStorage{db: db}, nil
}

func initDB(ctx context.Context, db DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id uuid NOT NULL,
			short_url text NOT NULL,
			original_url text NOT NULL,
			CONSTRAINT urls_pk PRIMARY KEY (id),
			CONSTRAINT urls_short_url_unique UNIQUE (short_url)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}

func (s *DBStorage) SaveShortURL(ctx context.Context, model models.URL) error {
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO urls (id, short_url, original_url)
        VALUES ($1, $2, $3)
        ON CONFLICT (short_url) DO NOTHING
    `, model.ID, model.ShortURL, model.OriginalURL)
	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}

	return nil
}

func (s *DBStorage) GetLongURL(ctx context.Context, short string) (string, bool) {
	var long string
	err := s.db.QueryRowContext(ctx, `
        SELECT original_url
        FROM urls
        WHERE short_url = $1
        LIMIT 1
    `, short).Scan(&long)

	if err != nil {
		return "", false
	}

	return long, true
}
