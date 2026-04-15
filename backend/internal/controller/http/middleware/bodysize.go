package middleware

import (
	"net/http"
	"strings"
)

// MaxBodySizeMiddleware ограничивает размер тела запроса для всех эндпоинтов,
// кроме multipart/form-data (загрузка файлов) — там хендлер сам устанавливает лимит.
func MaxBodySizeMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
