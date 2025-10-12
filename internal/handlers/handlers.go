package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/linarium/shortener/internal/handlers/middleware"
	"github.com/linarium/shortener/internal/logger"
	"github.com/linarium/shortener/internal/usecase"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/models"
)

const defaultContentType = "text/plain"

type URLHandler struct {
	shortener usecase.Repository
	config    config.Config
}

func NewURLHandler(cfg config.Config, shortener usecase.Repository) *URLHandler {
	if shortener == nil {
		panic("shortener cannot be nil")
	}
	return &URLHandler{
		shortener: shortener,
		config:    cfg,
	}
}

func (h *URLHandler) getUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}

func (h *URLHandler) buildShortURL(shortKey string) (string, error) {
	if h.config.BaseURL == "" {
		return "", fmt.Errorf("base URL is not configured")
	}
	fullURL, err := url.JoinPath(h.config.BaseURL, shortKey)
	if err != nil {
		return "", fmt.Errorf("failed to build short URL: %w", err)
	}
	return fullURL, nil
}

func (h *URLHandler) createJSONShortURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r.Context())
	if !ok {
		logger.Sugar.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var request struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	shortKey, isDuplicate := h.shortener.Shorten(r.Context(), request.URL, userID)

	shortURL, err := h.buildShortURL(shortKey)
	if err != nil {
		logger.Sugar.Errorf("Failed to build short URL: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := struct {
		Result string `json:"result"`
	}{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if isDuplicate {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Sugar.Errorf("Failed to encode response: %v", err)
		return
	}
}

func (h *URLHandler) createShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Incorrect method", http.StatusBadRequest)
		return
	}

	userID, ok := h.getUserID(r.Context())
	if !ok {
		logger.Sugar.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
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

	shortKey, isDuplicate := h.shortener.Shorten(r.Context(), input, userID)

	resultURL, err := h.buildShortURL(shortKey)
	if err != nil {
		logger.Sugar.Errorf("Failed to build short URL: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", defaultContentType)
	if isDuplicate {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	if _, err := w.Write([]byte(resultURL)); err != nil {
		logger.Sugar.Errorf("Failed to write response: %v", err)
		return
	}
}

func (h *URLHandler) getURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	url, exists, isDeleted := h.shortener.Expand(r.Context(), id)

	if !exists {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	if isDeleted {
		http.Error(w, "URL has been deleted", http.StatusGone)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *URLHandler) GetURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r.Context())
	if !ok {
		// Если userID нет в контексте - возвращаем 401 Unauthorized
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.shortener.GetUserURLs(r.Context(), userID)
	if err != nil {
		logger.Sugar.Errorf("failed to get user urls: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Преобразуем в нужный формат ответа
	response := make([]struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}, len(urls))

	for i, url := range urls {
		shortURL, err := h.buildShortURL(url.ShortURL)
		if err != nil {
			logger.Sugar.Errorf("Failed to build short URL: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response[i] = struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}{
			ShortURL:    shortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Sugar.Errorf("failed to encode response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.shortener.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *URLHandler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	logger.Sugar.Info("ShortenBatch called")

	userID, ok := h.getUserID(r.Context())
	if !ok {
		logger.Sugar.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Sugar.Infof("User ID: %s", userID)

	var req models.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	logger.Sugar.Infof("Received %d URLs to shorten", len(req))

	if len(req) == 0 {
		http.Error(w, "Empty batch request", http.StatusBadRequest)
		return
	}

	resp, err := h.shortener.ShortenBatch(r.Context(), req, h.config.BaseURL, userID)
	if err != nil {
		logger.Sugar.Errorf("Error in ShortenBatch: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Sugar.Errorf("Error encoding response: %v", err)
		http.Error(w, "Failed to prepare response", http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := h.getUserID(r.Context())
	if !ok {
		logger.Sugar.Warn("Unauthorized access to delete URLs")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		logger.Sugar.Errorf("Error decoding delete request: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(shortURLs) == 0 {
		logger.Sugar.Warn("Empty delete request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Sugar.Infof("User %s requested deletion of %d URLs", userID, len(shortURLs))

	go func() {
		if err := h.shortener.DeleteURLs(context.Background(), userID, shortURLs); err != nil {
			logger.Sugar.Errorf("Error deleting URLs: %v", err)
		} else {
			logger.Sugar.Infof("Successfully processed deletion of %d URLs for user %s", len(shortURLs), userID)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}
