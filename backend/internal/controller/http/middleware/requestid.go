package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware добавляет уникальный X-Request-ID к каждому запросу.
// Если клиент уже передал заголовок — используем его, иначе генерируем новый.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.New().String()
		}
		w.Header().Set(RequestIDHeader, reqID)
		next.ServeHTTP(w, r)
	})
}
