package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func InitConfig() Config {
	cfg := Config{}

	// Получаем значения из переменных окружения
	serverAddress := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Базовый адрес для сокращённого URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/shortener.json", "Путь до файла для сохранения данных")
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

	return cfg
}

// String возвращает строковое представление конфигурации
func (c *Config) String() string {
	return fmt.Sprintf("ServerAddress: %s, BaseURL: %s", c.ServerAddress, c.BaseURL)
}
