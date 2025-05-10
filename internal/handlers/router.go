package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/service"
	"net/http"
)

func Router(cfg config.Config, shortener *service.URLShortener) chi.Router {
	r := chi.NewRouter()

	handler := NewURLHandler(cfg, shortener)

	r.Use(WithLogging)

	r.Post("/", Compressor(handler.createShortURL))
	r.Post("/api/shorten", Compressor(handler.createJSONShortURL))
	r.Get("/{id}", Compressor(handler.getURL))
	r.Get("/ping", Compressor(handler.PingDB))

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	return r
}
