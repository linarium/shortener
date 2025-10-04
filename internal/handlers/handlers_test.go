package handlers

import (
	"context"
	"github.com/linarium/shortener/internal/handlers/middleware"
	"github.com/linarium/shortener/internal/usecase"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/linarium/shortener/internal/config"
	"github.com/linarium/shortener/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestCreateShortURL(t *testing.T) {
	cfg := config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		SecretKey:     "test-secret-key",
	}
	storage, _ := service.NewMemoryStorage(context.Background())
	shortener := usecase.NewShortenerService(storage)

	handler := NewURLHandler(cfg, shortener)

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		userID         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid request",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "http://example.com",
			userID:         "test-user-id",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			contentType:    "text/plain",
			body:           "http://example.com",
			userID:         "test-user-id",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "",
			userID:         "test-user-id",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Добавляем userID в контекст
			ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.createShortURL(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				if !strings.HasPrefix(body, tt.expectedBody) {
					t.Errorf("expected body to start with %s, got %s", tt.expectedBody, body)
				}
			}
		})
	}
}

func TestGetURL(t *testing.T) {
	cfg := config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		SecretKey:     "test-secret-key",
	}
	storage, _ := service.NewMemoryStorage(context.Background())
	shortener := usecase.NewShortenerService(storage)

	handler := NewURLHandler(cfg, shortener)

	// Сначала создаем URL с userID
	originalURL := "http://example.com"
	userID := "test-user-id"
	shortURL, _ := handler.shortener.Shorten(context.WithValue(context.Background(), middleware.UserIDContextKey, userID), originalURL, userID)

	tests := []struct {
		name           string
		method         string
		id             string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid ID",
			method:         http.MethodGet,
			id:             shortURL,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    originalURL,
		},
		{
			name:           "Invalid ID",
			method:         http.MethodGet,
			id:             "nonexistent",
			expectedStatus: http.StatusBadRequest,
			expectedURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/"+tt.id, nil)
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/{id}", handler.getURL)
			r.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedURL != "" {
				location := resp.Header.Get("Location")
				if location != tt.expectedURL {
					t.Errorf("expected Location header %s, got %s", tt.expectedURL, location)
				}
			}
		})
	}
}

func TestGetURLs(t *testing.T) {
	cfg := config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		SecretKey:     "test-secret-key",
	}
	storage, _ := service.NewMemoryStorage(context.Background())
	shortener := usecase.NewShortenerService(storage)

	handler := NewURLHandler(cfg, shortener)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "No user ID",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "User with no URLs",
			userID:         "user-with-no-urls",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "User with URLs",
			userID:         "user-with-urls",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "User with URLs" {
				ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, tt.userID)
				handler.shortener.Shorten(ctx, "http://example1.com", tt.userID)
				handler.shortener.Shorten(ctx, "http://example2.com", tt.userID)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

			// Добавляем userID в контекст только если он не пустой
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			handler.GetURLs(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestCreateJSONShortURL(t *testing.T) {
	cfg := config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		SecretKey:     "test-secret-key",
	}
	storage, _ := service.NewMemoryStorage(context.Background())
	shortener := usecase.NewShortenerService(storage)

	handler := NewURLHandler(cfg, shortener)

	tests := []struct {
		name           string
		body           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Valid JSON request",
			body:           `{"url": "http://example.com"}`,
			userID:         "test-user-id",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			body:           `{"url": }`,
			userID:         "test-user-id",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty URL",
			body:           `{"url": ""}`,
			userID:         "test-user-id",
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Добавляем userID в контекст
			ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.createJSONShortURL(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
