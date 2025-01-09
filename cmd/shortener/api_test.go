package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOrCreateUrl(t *testing.T) {
	tests := []struct {
		name           string
		urlPath        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Existing Short URL",
			urlPath:        "/EwHXdJfB",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedBody:   "",
		},
		{
			name:           "Non-Existing Short URL",
			urlPath:        "/nonexistent",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.urlPath, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			getOrCreateUrl(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestGetOrCreateUrl_POST(t *testing.T) {
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString("https://new-example.com"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	getOrCreateUrl(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	if !bytes.HasPrefix(rr.Body.Bytes(), []byte("http://localhost:8080/")) {
		t.Errorf("handler returned unexpected body: got %v want prefix %v",
			rr.Body.String(), "http://localhost:8080/")
	}
}
