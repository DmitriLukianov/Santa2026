package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// otpEntry хранит счётчик и время начала окна для одного ключа (email или IP).
type otpEntry struct {
	count     int
	windowEnd time.Time
}

// OTPRateLimiter — простой in-memory ограничитель числа OTP-запросов.
// Использует скользящее окно в 1 час. Безопасен для конкурентного доступа.
type OTPRateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*otpEntry
	maxPerHr int
}

func NewOTPRateLimiter(maxPerHour int) *OTPRateLimiter {
	rl := &OTPRateLimiter{
		entries:  make(map[string]*otpEntry),
		maxPerHr: maxPerHour,
	}
	// Фоновая очистка устаревших записей каждые 10 минут
	go rl.cleanup()
	return rl
}

func (rl *OTPRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	e, ok := rl.entries[key]
	if !ok || now.After(e.windowEnd) {
		rl.entries[key] = &otpEntry{count: 1, windowEnd: now.Add(time.Hour)}
		return true
	}
	if e.count >= rl.maxPerHr {
		return false
	}
	e.count++
	return true
}

func (rl *OTPRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, e := range rl.entries {
			if now.After(e.windowEnd) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

// OTPRateLimitMiddleware возвращает middleware, ограничивающий запросы по email из тела запроса.
// Если email недоступен — ограничивает по IP.
func OTPRateLimitMiddleware(rl *OTPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Определяем ключ: пробуем извлечь email из query или используем IP
			key := r.RemoteAddr
			if email := r.URL.Query().Get("email"); email != "" {
				key = "email:" + email
			}

			if !rl.Allow(key) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "3600")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Слишком много запросов. Попробуйте через час.",
					"code":  http.StatusTooManyRequests,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
