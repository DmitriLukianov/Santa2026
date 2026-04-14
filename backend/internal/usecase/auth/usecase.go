package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/dto"
	"secret-santa-backend/internal/oauth"
	"secret-santa-backend/internal/usecase"
)

type UseCase struct {
	userUC           usecase.UserUseCase
	emailService     usecase.EmailService
	verificationRepo usecase.VerificationRepository
	smtpEnabled      bool
	log              *slog.Logger
}

func New(userUC usecase.UserUseCase, emailService usecase.EmailService, verificationRepo usecase.VerificationRepository, smtpEnabled bool) *UseCase {
	return &UseCase{
		userUC:           userUC,
		emailService:     emailService,
		verificationRepo: verificationRepo,
		smtpEnabled:      smtpEnabled,
	}
}

func NewWithLogger(userUC usecase.UserUseCase, emailService usecase.EmailService, verificationRepo usecase.VerificationRepository, smtpEnabled bool, log *slog.Logger) *UseCase {
	return &UseCase{
		userUC:           userUC,
		emailService:     emailService,
		verificationRepo: verificationRepo,
		smtpEnabled:      smtpEnabled,
		log:              log,
	}
}

func (uc *UseCase) LoginWithOAuth(ctx context.Context, info oauth.UserInfo) (string, error) {
	if uc.log != nil {
		uc.log.Info("oauth login started",
			slog.String("provider", info.Provider),
			slog.String("oauth_id", info.ID),
		)
	}

	if info.ID == "" {
		return "", definitions.ErrMissingOAuthCode
	}

	user, err := uc.userUC.GetByOAuthID(ctx, info.ID, info.Provider)
	if err == nil && user != nil {
		if uc.log != nil {
			uc.log.Info("existing oauth user found", slog.String("user_id", user.ID.String()))
		}
		if uc.emailService != nil {
			go func() { _ = uc.emailService.SendLoginNotification(context.Background(), user.Email, user.Name) }()
		}
		return user.ID.String(), nil
	}

	if !errors.Is(err, definitions.ErrUserNotFound) && err != nil {
		return "", fmt.Errorf("failed to lookup oauth user: %w", err)
	}

	// Try to find existing account by email (e.g. registered via OTP earlier)
	if info.Email != "" {
		existingUser, emailErr := uc.userUC.GetByEmail(ctx, info.Email)
		if emailErr == nil && existingUser != nil {
			if uc.log != nil {
				uc.log.Info("oauth user matched existing email account",
					slog.String("user_id", existingUser.ID.String()),
					slog.String("email", info.Email),
				)
			}
			if uc.emailService != nil {
				go func() {
					_ = uc.emailService.SendLoginNotification(context.Background(), existingUser.Email, existingUser.Name)
				}()
			}
			return existingUser.ID.String(), nil
		}
	}

	createInput := dto.CreateUserInput{
		Name:          info.Name,
		Email:         info.Email,
		OAuthID:       info.ID,
		OAuthProvider: info.Provider,
	}
	createdUser, err := uc.userUC.Create(ctx, createInput)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	if uc.log != nil {
		uc.log.Info("new oauth user created", slog.String("user_id", createdUser.ID.String()))
	}
	if uc.emailService != nil {
		go func() { _ = uc.emailService.SendLoginNotification(context.Background(), createdUser.Email, createdUser.Name) }()
	}
	return createdUser.ID.String(), nil
}

func (uc *UseCase) SendOTP(ctx context.Context, email string) error {
	if uc.log != nil {
		uc.log.Info("send otp started", slog.String("email", email))
	}

	code, err := uc.emailService.SendOTP(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	// Когда SMTP не настроен, email не отправляется, но код сохраняется.
	// Логируем код, чтобы разработчик мог использовать его вручную.
	if !uc.smtpEnabled && uc.log != nil {
		uc.log.Warn("SMTP не настроен — OTP-код не отправлен на почту",
			slog.String("email", email),
			slog.String("otp_code", code),
		)
	}

	// Инвалидируем все старые коды для этого email перед отправкой нового
	_ = uc.verificationRepo.InvalidateCodes(ctx, email)

	expiresAt := time.Now().Add(10 * time.Minute)
	if err := uc.verificationRepo.SaveCode(ctx, email, code, expiresAt); err != nil {
		return fmt.Errorf("failed to save verification code: %w", err)
	}

	return nil
}

func (uc *UseCase) VerifyOTP(ctx context.Context, email, code, name string) (string, error) {
	if uc.log != nil {
		uc.log.Info("verify otp started", slog.String("email", email))
	}

	valid, err := uc.verificationRepo.GetValidCode(ctx, email, code)
	if err != nil || !valid {
		return "", definitions.ErrInvalidUserInput
	}

	if err := uc.verificationRepo.MarkAsUsed(ctx, email, code); err != nil {
		if uc.log != nil {
			uc.log.Warn("failed to mark OTP as used", slog.String("email", email), slog.String("error", err.Error()))
		}
		return "", definitions.ErrInvalidUserInput
	}

	user, err := uc.userUC.GetByEmail(ctx, email)
	if err == nil && user != nil {
		if uc.emailService != nil {
			go func() { _ = uc.emailService.SendLoginNotification(context.Background(), user.Email, user.Name) }()
		}
		return user.ID.String(), nil
	}

	defaultName := nameFromEmail(email)
	if name != "" {
		defaultName = name
	}

	createInput := dto.CreateUserInput{
		Name:          defaultName,
		Email:         email,
		OAuthID:       email,
		OAuthProvider: "email",
	}
	createdUser, err := uc.userUC.Create(ctx, createInput)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	if uc.emailService != nil {
		go func() { _ = uc.emailService.SendLoginNotification(context.Background(), createdUser.Email, createdUser.Name) }()
	}
	return createdUser.ID.String(), nil
}

func nameFromEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "Пользователь"
	}
	return parts[0]
}
