package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/service"
	"io"
	"net/http"
)

const defaultContentType = "text/plain"

type URLHandler struct {
	shortener *service.URLShortener
	config    config.Config
}

func NewURLHandler(cfg config.Config, shortener *service.URLShortener) *URLHandler {
	return &URLHandler{
		shortener: shortener,
		config:    cfg,
	}
}

func (h *URLHandler) createJSONShortURL(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	shortURL := h.shortener.Shorten(request.URL)

	response := struct {
		Result string `json:"result"`
	}{
		Result: h.config.BaseURL + "/" + shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *URLHandler) createShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Incorrect method", http.StatusBadRequest)
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

	var input string

	input = string(body)
	w.Header().Set("Content-Type", defaultContentType)
	shortURL := h.shortener.Shorten(input)
	host := h.config.BaseURL
	resultURL := host + "/" + shortURL

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (h *URLHandler) getURL(w http.ResponseWriter, r *http.Request) {
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

func Router(cfg config.Config, shortener *service.URLShortener) chi.Router {
	r := chi.NewRouter()

	handler := NewURLHandler(cfg, shortener)

	r.Use(WithLogging)

	r.Post("/", handler.createShortURL)
	r.Post("/api/shorten", handler.createJSONShortURL)
	r.Get("/{id}", handler.getURL)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	return r
}
