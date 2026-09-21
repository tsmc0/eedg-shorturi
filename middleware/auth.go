package middleware

import (
	"crypto/subtle"
	"net/http"

	"eedg-shortlink/helpers"
)

// APIKey возвращает middleware для проверки X-API-Key.
// Если key пустой — авторизация отключена.
func APIKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			got := r.Header.Get("X-API-Key")
			if subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
				helpers.WriteError(w, http.StatusUnauthorized, "invalid api key")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
