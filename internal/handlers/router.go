package handlers

import (
	"github.com/linarium/shortener/internal/handlers/middleware"
	"github.com/linarium/shortener/internal/usecase"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linarium/shortener/internal/config"
)

func Router(cfg config.Config, shortener usecase.Repository) chi.Router {
	r := chi.NewRouter()

	handler := NewURLHandler(cfg, shortener)

	r.Use(middleware.Authenticate(cfg.SecretKey))
	r.Use(middleware.WithLogging)

	r.Get("/{id}", middleware.Compressor(handler.getURL))
	r.Get("/ping", middleware.Compressor(handler.PingDB))

	r.Group(func(r chi.Router) {
		r.Post("/", middleware.Compressor(handler.createShortURL))
		r.Post("/api/shorten", middleware.Compressor(handler.createJSONShortURL))
		r.Post("/api/shorten/batch", handler.ShortenBatch)
		r.Get("/api/user/urls", handler.GetURLs)
		r.Delete("/api/user/urls", handler.DeleteURLs)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	return r
}
