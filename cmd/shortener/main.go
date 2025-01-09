package main

import (
	"fmt"
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/handlers"
	"net/http"
)

func main() {
	cfg := config.InitConfig()
	fmt.Println("Configuration:", cfg)

	r := handlers.Router()
	fmt.Println("Server is running on port 8080...")
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
