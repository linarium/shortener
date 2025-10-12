package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/linarium/shortener/internal/logger"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const UserIDContextKey contextKey = "userID"
const cookieName = "user_id"

func Authenticate(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string

			cookie, err := r.Cookie(cookieName)
			if err != nil {
				// Куки нет, создаем нового пользователя
				userID = generateUserID()
				setAuthCookie(w, userID, secretKey)
				logger.Sugar.Debugf("Created new user ID: %s", userID)
			} else {
				// Проверяем подпись куки
				userID, err = validateCookie(cookie, secretKey)
				if err != nil {
					// Кука невалидна, создаем нового пользователя
					userID = generateUserID()
					setAuthCookie(w, userID, secretKey)
					logger.Sugar.Debugf("Invalid cookie, created new user ID: %s", userID)
				} else {
					logger.Sugar.Debugf("Authenticated user ID: %s", userID)
				}
			}

			// Добавляем userID в контекст
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// generateUserID создает новый UUID для пользователя
func generateUserID() string {
	return uuid.New().String()
}

// signData создает подпись для данных
func signData(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// setAuthCookie устанавливает аутентификационную куку
func setAuthCookie(w http.ResponseWriter, userID, secretKey string) {
	signature := signData(userID, secretKey)
	value := userID + "." + signature

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour), // 1 год
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

// validateCookie проверяет валидность куки
func validateCookie(cookie *http.Cookie, secretKey string) (string, error) {
	if cookie.Value == "" {
		return "", errors.New("empty cookie value")
	}

	// Разделяем userID и подпись
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return "", errors.New("invalid cookie format")
	}

	userID := parts[0]
	signature := parts[1]

	// Проверяем что userID - валидный UUID
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Проверяем подпись
	expectedSignature := signData(userID, secretKey)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", errors.New("invalid signature")
	}

	return userID, nil
}

// GetUserIDFromContext извлекает userID из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}
