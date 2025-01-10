package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCreateShortURL(t *testing.T) {
	handler := NewURLHandler()

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid request",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "http://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			contentType:    "text/plain",
			body:           "http://example.com",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "Invalid content type",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           "http://example.com",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
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
	handler := NewURLHandler()

	originalURL := "http://example.com"
	shortURL := handler.shortener.Shorten(originalURL)

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
