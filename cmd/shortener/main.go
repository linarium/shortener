package main

import (
	"context"
	"net/http"

	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/service"

	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		logger.Sugar.Fatalf("Ошибка инициализации конфигурации: %v\n", err)
	}

	logger.Initialize()
	defer logger.Sugar.Sync()

	storage, err := service.NewDBStorage(context.Background(), "pgx", cfg.DatabaseDSN)
	if err != nil {
		logger.Sugar.Fatalf("Ошибка при создании хранилища: %v", err)
	}
	defer storage.Close()

	shortener := service.NewURLShortener(storage)

	r := handlers.Router(cfg, shortener)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		logger.Sugar.Fatalf("Сбой в работе сервера: %v", err)
	}
}
