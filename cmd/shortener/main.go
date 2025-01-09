package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

// Хранилище сокращённых URL
var urlStorage = map[string]string{
	"EwHXdJfB": "https://practicum.yandex.ru/",
	"abc123":   "https://example.com/",
}

// Генерация случайного идентификатора для сокращённого URL
func generateID() string {
	b := make([]byte, 6) // Генерируем 6 случайных байт
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:8]
}

// Обработчик для POST-запросов
func createShortURL(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
	}
	defer r.Body.Close()

	originalURL := string(body)
	// Генерируем уникальный идентификатор
	id := generateID()
	// Сохраняем оригинальный URL в хранилище
	urlStorage[id] = originalURL

	// Формируем сокращённый URL
	shortURL := fmt.Sprintf("http://localhost:8080/%s", id)

	// Устанавливаем заголовки и возвращаем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func getOrCreateUrl(w http.ResponseWriter, r *http.Request) {
	// Если POST запрос - запускаем соответствующую функцию, для создания сокращённого URL
	if r.Method == http.MethodPost {
		createShortURL(w, r)
		return
	}

	// Извлекаем id из пути запроса
	id := r.URL.Path[1:]

	// Проверяем, существует ли сокращённый URL в хранилище
	if originalURL, exists := urlStorage[id]; exists {
		// Устанавливаем заголовки
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getOrCreateUrl)

	fmt.Println("Server is running on port 8080...")
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
