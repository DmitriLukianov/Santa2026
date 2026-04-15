package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string `env:"APP_PORT" envDefault:"8080"`
	AppEnv     string `env:"APP_ENV" envDefault:"local"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info"`
	AppBaseURL  string `env:"APP_BASE_URL" envDefault:"http://localhost:8080"`
	FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:5173"`

	// CORSOrigins — список разрешённых источников через запятую.
	// Пример: "https://myapp.com,https://www.myapp.com"
	// По умолчанию разрешает localhost для разработки.
	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173,http://localhost:3000"`

	DatabaseURL string `env:"DATABASE_URL"`

	// Настройки пула соединений БД
	DBMaxConns     int `env:"DB_MAX_CONNS" envDefault:"25"`
	DBMinConns     int `env:"DB_MIN_CONNS" envDefault:"5"`

	JWTSecret string        `env:"JWT_SECRET"`
	JWTTTL    time.Duration `env:"JWT_TTL" envDefault:"24h"`

	// SMTP — опциональный. Если не задан, email-уведомления отключаются.
	// Сервис стартует без них. Проверяй через cfg.SMTPEnabled().
	SMTPHost     string `env:"SMTP_HOST" envDefault:"smtp.mail.ru"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	FromEmail    string `env:"FROM_EMAIL"`

	OAuthProvider      string `env:"OAUTH_PROVIDER" envDefault:"github"`
	GithubClientID     string `env:"GITHUB_CLIENT_ID"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET"`
	GithubRedirectURL  string `env:"GITHUB_REDIRECT_URL"`

	// UploadDir — директория для загружаемых файлов.
	UploadDir string `env:"UPLOAD_DIR" envDefault:"./uploads"`

	// MaxRequestBodySize — максимальный размер тела запроса (байты), кроме загрузки файлов.
	MaxRequestBodySize int64 `env:"MAX_REQUEST_BODY_SIZE" envDefault:"1048576"` // 1 MB

	// OTPLength — длина OTP-кода.
	OTPLength int `env:"OTP_LENGTH" envDefault:"6"`

	// OTPExpiryMinutes — время жизни OTP-кода в минутах.
	OTPExpiryMinutes int `env:"OTP_EXPIRY_MINUTES" envDefault:"10"`

	// RateLimitOTPPerHour — максимальное число запросов OTP с одного email в час.
	RateLimitOTPPerHour int `env:"RATE_LIMIT_OTP_PER_HOUR" envDefault:"5"`
}

// SMTPEnabled возвращает true если все необходимые SMTP-переменные заданы.
// Используй это перед отправкой писем вместо падения сервиса.
func (c *Config) SMTPEnabled() bool {
	return c.SMTPHost != "" && c.SMTPUsername != "" && c.SMTPPassword != "" && c.FromEmail != ""
}

// CORSOriginsSlice возвращает список разрешённых CORS-источников.
func (c *Config) CORSOriginsSlice() []string {
	var origins []string
	for _, o := range splitAndTrim(c.CORSOrigins) {
		if o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:5173"}
	}
	return origins
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range splitComma(s) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		AppEnv:     getEnv("APP_ENV", "local"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		AppBaseURL:  getEnv("APP_BASE_URL", "http://localhost:8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		DBMaxConns: getIntEnv("DB_MAX_CONNS", 25),
		DBMinConns: getIntEnv("DB_MIN_CONNS", 5),

		JWTSecret: getEnv("JWT_SECRET", ""),
		JWTTTL:    parseDuration(getEnv("JWT_TTL", "24h")),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.mail.ru"),
		SMTPPort:     getIntEnv("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),

		OAuthProvider:      getEnv("OAUTH_PROVIDER", "github"),
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GithubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),

		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
		MaxRequestBodySize: getInt64Env("MAX_REQUEST_BODY_SIZE", 1<<20),
		OTPLength:          getIntEnv("OTP_LENGTH", 6),
		OTPExpiryMinutes:   getIntEnv("OTP_EXPIRY_MINUTES", 10),
		RateLimitOTPPerHour: getIntEnv("RATE_LIMIT_OTP_PER_HOUR", 5),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}

	if !cfg.SMTPEnabled() {
		log.Println("WARNING: SMTP not configured — email notifications are disabled")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getInt64Env(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("Warning: invalid JWT_TTL '%s', using default 24h", s)
		return 24 * time.Hour
	}
	return d
}
