package main

import (
	"github.com/linarium/shortener/internal/logger"
	"net/http"

	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
)

func main() {
	cfg := config.InitConfig()

	logger.Initialize()
	defer logger.Sugar.Sync()

	r := handlers.Router(cfg)
	err := http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		logger.Sugar.Fatalf("Сбой в работе сервера: %v", err)
	}
}
