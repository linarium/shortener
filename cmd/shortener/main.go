package main

import (
	"log"
	"net/http"

	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
)

func main() {
	cfg := config.InitConfig()

	r := handlers.Router(cfg)
	err := http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Fatalf("Сбой в работе сервера: %v", err)
	}
}
