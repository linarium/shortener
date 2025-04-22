package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/models"
	"github.com/linarium/shortener/internal/service"
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

	shortKey, isDuplicate := h.shortener.Shorten(r.Context(), request.URL)

	response := struct {
		Result string `json:"result"`
	}{
		Result: h.config.BaseURL + "/" + shortKey,
	}

	w.Header().Set("Content-Type", "application/json")
	if isDuplicate {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
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

	input := string(body)
	w.Header().Set("Content-Type", defaultContentType)

	shortURL, isDuplicate := h.shortener.Shorten(r.Context(), input)
	resultURL := h.config.BaseURL + "/" + shortURL

	w.Header().Set("Content-Type", defaultContentType)
	if isDuplicate {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write([]byte(resultURL))
}

func (h *URLHandler) getURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	url, exists := h.shortener.Expand(r.Context(), id)

	if !exists {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *URLHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	db, err := service.NewDB(h.config.DatabaseDSN)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := db.PingContext(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *URLHandler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	var req models.BatchRequest
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, "Empty batch request", http.StatusBadRequest)
		return
	}

	resp, err := h.shortener.ShortenBatch(r.Context(), req, h.config.BaseURL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Println(err)
	}
}
