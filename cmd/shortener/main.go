package main

import (
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
	"net/http"
)

func main() {
	cfg := config.InitConfig()

	r := handlers.Router(cfg)
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
