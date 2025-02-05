package service

import (
	"bufio"
	"encoding/json"
	"github.com/linarium/shortener/internal/models"
	"os"
	"sync"
)

type Storage interface {
	SaveShortURL(model models.URL) error
	GetLongURL(short string) (string, bool)
	Close() error
}

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{data: make(map[string]string)}, nil
}

func (s *MemoryStorage) SaveShortURL(model models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[model.ShortURL] = model.OriginalURL
	return nil
}

func (s *MemoryStorage) GetLongURL(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.data[short]
	return long, exists
}

func (s *MemoryStorage) Close() error {
	return nil
}

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

	return &FileStorage{
		file:   file,
		writer: bufio.NewWriter(file),
		memory: MemoryStorage{data: data},
	}, nil
}

func (s *FileStorage) GetLongURL(short string) (string, bool) {
	return s.memory.GetLongURL(short)
}

func (s *FileStorage) SaveShortURL(model models.URL) error {
	if err := json.NewEncoder(s.writer).Encode(model); err != nil {
		return err
	}
	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.SaveShortURL(model)
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}
