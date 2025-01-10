package main

import (
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
	"net/http"
)

func main() {
	config.InitConfig()

	r := handlers.Router()
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
