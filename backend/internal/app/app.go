package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"secret-santa-backend/internal/config"
	"secret-santa-backend/internal/database"
	"secret-santa-backend/internal/email"
	"secret-santa-backend/internal/logger"
	"secret-santa-backend/internal/oauth"
	"secret-santa-backend/internal/scheduler"

	assignmentrepo "secret-santa-backend/internal/repository/postgres/assignment"
	chatrepo "secret-santa-backend/internal/repository/postgres/chat"
	eventrepo "secret-santa-backend/internal/repository/postgres/event"
	invitationrepo "secret-santa-backend/internal/repository/postgres/invitation"
	participantrepo "secret-santa-backend/internal/repository/postgres/participant"
	userrepo "secret-santa-backend/internal/repository/postgres/user"
	verificationrepo "secret-santa-backend/internal/repository/postgres/verification"
	wishlistrepo "secret-santa-backend/internal/repository/postgres/wishlist"

	assignmentusecase "secret-santa-backend/internal/usecase/assignment"
	authusecase "secret-santa-backend/internal/usecase/auth"
	chatusecase "secret-santa-backend/internal/usecase/chat"
	eventusecase "secret-santa-backend/internal/usecase/event"
	invitationusecase "secret-santa-backend/internal/usecase/invitation"
	participantusecase "secret-santa-backend/internal/usecase/participant"
	userusecase "secret-santa-backend/internal/usecase/user"
	wishlistusecase "secret-santa-backend/internal/usecase/wishlist"

	v1 "secret-santa-backend/internal/controller/http/v1"
)

type App struct {
	cfg    *config.Config
	log    *slog.Logger
	server *http.Server
	cancel context.CancelFunc
}

func New() *App {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel, cfg.AppEnv)

	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		panic(err)
	}

	userRepo := userrepo.New(db)
	eventRepo := eventrepo.New(db)
	participantRepo := participantrepo.New(db)
	assignmentRepo := assignmentrepo.New(db)
	wishlistRepo := wishlistrepo.New(db)
	invitationRepo := invitationrepo.New(db)
	chatRepo := chatrepo.New(db)
	verificationRepo := verificationrepo.New(db)
	userUC := userusecase.NewWithLogger(userRepo, log)
	emailService := email.New(cfg, log)

	authUC := authusecase.NewWithLogger(userUC, emailService, verificationRepo, cfg.SMTPEnabled(), log)

	eventUC := eventusecase.NewWithLogger(eventRepo, participantRepo, log)
	participantUC := participantusecase.NewWithLogger(participantRepo, eventRepo, log)
	assignmentUC := assignmentusecase.NewWithLogger(assignmentRepo, participantRepo, eventRepo, userUC, emailService, log)
	wishlistUC := wishlistusecase.NewWithLogger(wishlistRepo, participantRepo, assignmentRepo, log)
	invitationUC := invitationusecase.NewWithLogger(invitationRepo, eventRepo, participantUC, emailService, cfg.FrontendURL, log)
	chatUC := chatusecase.NewWithLogger(chatRepo, participantRepo, assignmentRepo, log)

	userHandler := v1.NewUserHandler(userUC)
	eventHandler := v1.NewEventHandler(eventUC)
	participantHandler := v1.NewParticipantHandler(participantUC, eventUC)
	wishlistHandler := v1.NewWishlistHandler(wishlistUC, participantUC)
	assignmentHandler := v1.NewAssignmentHandler(assignmentUC)
	invitationHandler := v1.NewInvitationHandler(invitationUC)
	chatHandler := v1.NewChatHandler(chatUC)

	jwtManager, err := oauth.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil {
		log.Error("failed to create JWT manager", slog.String("error", err.Error()))
		panic(err)
	}

	authProvider, err := oauth.New(cfg)
	if err != nil {
		log.Error("failed to create oauth provider", slog.String("error", err.Error()))
		panic(err)
	}

	authHandler := v1.NewAuthHandler(authProvider, jwtManager, authUC, userUC, cfg.FrontendURL)
	uploadHandler := v1.NewUploadHandler(cfg.AppBaseURL)

	router := v1.NewRouter(
		authHandler,
		userHandler,
		eventHandler,
		participantHandler,
		assignmentHandler,
		wishlistHandler,
		invitationHandler,
		chatHandler,
		uploadHandler,
		jwtManager,
		log,
	)

	ctx, cancel := context.WithCancel(context.Background())
	scheduler.New(eventRepo, assignmentUC, log).Start(ctx)

	return &App{
		cfg:    cfg,
		log:    log,
		cancel: cancel,
		server: &http.Server{
			Addr:         ":" + cfg.AppPort,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (a *App) Run() error {
	a.log.Info("Server running on :" + a.cfg.AppPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-quit:
		a.log.Info("shutting down server...")
		a.cancel()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return a.server.Shutdown(ctx)
	}
}
