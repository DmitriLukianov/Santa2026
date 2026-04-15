package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"secret-santa-backend/internal/controller/http/v1/response"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return RecoveryMiddlewareWithLogger(nil)(next)
}

func RecoveryMiddlewareWithLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					err := fmt.Errorf("panic: %v", rec)
					if log != nil {
						log.Error("recovered from panic",
							slog.String("error", fmt.Sprintf("%v", rec)),
							slog.String("stack", string(stack)),
							slog.String("path", r.URL.Path),
							slog.String("method", r.Method),
						)
					}
					response.WriteHTTPError(w, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
