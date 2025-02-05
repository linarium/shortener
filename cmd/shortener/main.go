package main

import (
	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/service"
	"net/http"

	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
)

func main() {
	cfg := config.InitConfig()

	logger.Initialize()
	defer logger.Sugar.Sync()

	storage, err := service.NewFileStorage(cfg.FileStoragePath)
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
