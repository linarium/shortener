package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func InitConfig() (Config, error) {
	cfg := Config{}

	// Получаем значения из переменных окружения
	serverAddress := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	databaseDSN := os.Getenv("DATABASE_DSN")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Базовый адрес для сокращённого URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "/data/shortener.json", "Путь до файла для сохранения данных")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN")
	flag.Parse()

	// Приоритет: переменные окружения > флаги > значения по умолчанию
	if serverAddress != "" {
		cfg.ServerAddress = serverAddress
	}
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if fileStoragePath != "" {
		cfg.FileStoragePath = fileStoragePath
	}
	if databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validateConfig(cfg Config) error {
	if cfg.ServerAddress == "" {
		return fmt.Errorf("ServerAddress не может быть пустым")
	}

	if cfg.BaseURL == "" {
		return fmt.Errorf("BaseURL не может быть пустым")
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return fmt.Errorf("BaseURL должен быть корректным URL: %v", err)
	}

	if cfg.FileStoragePath == "" {
		return fmt.Errorf("FileStoragePath не может быть пустым")
	}
	if !filepath.IsAbs(cfg.FileStoragePath) {
		return fmt.Errorf("FileStoragePath должен быть абсолютным путём")
	}

	return nil
}

// String возвращает строковое представление конфигурации
func (c *Config) String() string {
	return fmt.Sprintf("ServerAddress: %s, BaseURL: %s", c.ServerAddress, c.BaseURL)
}
