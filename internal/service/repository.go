package service

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/linarium/shortener/internal/models"
)

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	memory *MemoryStorage
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
		memory: &MemoryStorage{data: data},
	}, nil
}

func (s *FileStorage) GetLongURL(ctx context.Context, short string) (string, bool, bool) {
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

func (s *FileStorage) SaveManyURLS(ctx context.Context, models []models.URL) error {
	for _, model := range models {
		if err := json.NewEncoder(s.writer).Encode(model); err != nil {
			return err
		}
	}

	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.SaveManyURLS(ctx, models)
}

func (s *FileStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	return s.memory.GetAll(ctx, userID)
}

func (s *FileStorage) DeleteURLs(ctx context.Context, userID string, shortURLs []string) error {
	return s.memory.DeleteURLs(ctx, userID, shortURLs)
}
