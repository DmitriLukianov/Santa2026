package v1

import (
	"log/slog"
	"net/http"
	"time"

	"secret-santa-backend/internal/config"
	"secret-santa-backend/internal/controller/http/middleware"
	"secret-santa-backend/internal/oauth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	eventHandler *EventHandler,
	participantHandler *ParticipantHandler,
	assignmentHandler *AssignmentHandler,
	wishlistHandler *WishlistHandler,
	invitationHandler *InvitationHandler,
	chatHandler *ChatHandler,
	uploadHandler *UploadHandler,
	jwtManager *oauth.JWTManager,
	log *slog.Logger,
	cfg *config.Config,
	db *pgxpool.Pool,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOriginsSlice(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Use(middleware.RecoveryMiddlewareWithLogger(log))
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.MaxBodySizeMiddleware(cfg.MaxRequestBodySize))
	router.Use(middleware.TimeoutMiddleware(10 * time.Second))

	// Rate limiter для OTP-эндпоинтов
	otpLimiter := middleware.NewOTPRateLimiter(cfg.RateLimitOTPPerHour)

	router.Mount("/health", newHealthHandler(db))

	// Раздача загруженных файлов
	router.Handle("/static/uploads/*", http.StripPrefix("/static/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", authHandler.Login)
			r.Get("/callback", authHandler.Callback)
			// OTP-эндпоинты защищены rate limiter'ом
			r.With(middleware.OTPRateLimitMiddleware(otpLimiter)).Post("/send-otp", authHandler.SendOTP)
			r.With(middleware.OTPRateLimitMiddleware(otpLimiter)).Post("/verify-otp", authHandler.VerifyOTP)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.NewAuthMiddleware(jwtManager, log).Handler)

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", userHandler.GetMe)
				r.Patch("/me", userHandler.UpdateMe)
			})

			r.Route("/events", func(r chi.Router) {
				r.Post("/", eventHandler.CreateEvent)
				r.Get("/", eventHandler.GetEvents)
				r.Get("/{id}", eventHandler.GetEventByID)
				r.Put("/{id}", eventHandler.UpdateEvent)
				r.Delete("/{id}", eventHandler.DeleteEvent)
				r.Post("/{id}/finish", eventHandler.FinishEvent)

				r.Post("/{eventId}/participants", participantHandler.Add)
				r.Get("/{eventId}/participants", participantHandler.GetByEvent)

				r.Post("/{eventId}/assign", assignmentHandler.Draw)
				r.Get("/{eventId}/assignments", assignmentHandler.GetByEvent)

				r.Post("/{eventId}/open-invitation", invitationHandler.GenerateInviteByEvent)

				r.Route("/{eventId}/chat", func(r chi.Router) {
					r.Get("/recipient", chatHandler.GetRecipientChat)
					r.Get("/sender", chatHandler.GetSenderChat)
					r.Post("/messages", chatHandler.SendMessage)
					r.Post("/messages/santa", chatHandler.SendMessageToSanta)
				})
			})

			r.Delete("/participants/{id}", participantHandler.Delete)

			r.Route("/users/me/wishlist", func(r chi.Router) {
				r.Post("/", wishlistHandler.Create)
				r.Get("/", wishlistHandler.GetByUser)
			})

			r.Route("/wishlists/{participantId}", func(r chi.Router) {
				r.Get("/", wishlistHandler.GetByParticipant)
			})

			r.Route("/wishlists/{wishlistId}/items", func(r chi.Router) {
				r.Post("/", wishlistHandler.AddItem)
				r.Get("/", wishlistHandler.GetItems)
				r.Put("/{itemId}", wishlistHandler.UpdateItem)
				r.Delete("/{itemId}", wishlistHandler.DeleteItem)
			})

			r.Route("/invitations", func(r chi.Router) {
				r.Post("/generate", invitationHandler.GenerateInvite)
				r.Post("/send-email", invitationHandler.SendEmailInvitation)
			})
			r.Post("/invite/join", invitationHandler.JoinByInvite)

			r.Post("/upload", uploadHandler.Upload)
		})
	})

	return router
}
