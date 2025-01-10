package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/linarium/shortener/internal/service"
	"io"
	"net/http"
	"strings"
)

type URLHandler struct {
	shortener *service.URLShortener
}

func NewURLHandler() *URLHandler {
	return &URLHandler{shortener: service.NewURLShortener()}
}

func (h *URLHandler) createShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Incorrect method", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		http.Error(w, "Incorrect content type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL := h.shortener.Shorten(string(body))

	resultURL := fmt.Sprintf("http://localhost:8080/%s", shortURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (h *URLHandler) getURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Incorrect method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	url, exists := h.shortener.Expand(id)

	if !exists {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Router() chi.Router {
	r := chi.NewRouter()
	handler := NewURLHandler()

	r.Post("/", handler.createShortURL)
	r.Get("/{id}", handler.getURL)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	return r
}
