package main

import (
	"context"
	"github.com/linarium/shortener/internal/usecase"
	"log"
	"net/http"

	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/service"

	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Ошибка инициализации конфигурации: %v\n", err)
	}

	logger.Initialize()
	defer logger.Sync()

	storage, err := service.NewStorage(context.Background(), cfg)
	if err != nil {
		logger.Sugar.Fatalf("Ошибка при создании хранилища: %v", err)
	}
	defer storage.Close()

	shortener := usecase.NewShortenerService(storage)
	r := handlers.Router(cfg, shortener)

	logger.Sugar.Infof("Server starting on %s", cfg.ServerAddress)

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		logger.Sugar.Fatalf("Сбой в работе сервера: %v", err)
	}
}
