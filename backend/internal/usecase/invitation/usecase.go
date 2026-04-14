package invitation

import (
	"context"
	"fmt"
	"log/slog"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/dto"
	"secret-santa-backend/internal/entity"
	"secret-santa-backend/internal/usecase"

	"github.com/google/uuid"
)

type UseCase struct {
	repo           Repository
	eventRepo      EventRepository
	participantUC  usecase.ParticipantUseCase
	emailService   usecase.EmailService
	baseURL        string
	log            *slog.Logger
}

func New(repo Repository, eventRepo EventRepository, participantUC usecase.ParticipantUseCase, baseURL string) *UseCase {
	return &UseCase{
		repo:          repo,
		eventRepo:     eventRepo,
		participantUC: participantUC,
		baseURL:       baseURL,
	}
}

func NewWithLogger(repo Repository, eventRepo EventRepository, participantUC usecase.ParticipantUseCase, emailService usecase.EmailService, baseURL string, log *slog.Logger) *UseCase {
	return &UseCase{
		repo:          repo,
		eventRepo:     eventRepo,
		participantUC: participantUC,
		emailService:  emailService,
		baseURL:       baseURL,
		log:           log,
	}
}

func (uc *UseCase) GenerateInvite(ctx context.Context, input dto.CreateInvitationInput, organizerID uuid.UUID) (dto.InvitationResponse, error) {
	if uc.log != nil {
		uc.log.Info("generate invitation started",
			slog.String("event_id", input.EventID.String()),
			slog.String("organizer_id", organizerID.String()),
		)
	}

	event, err := uc.eventRepo.GetByID(ctx, input.EventID)
	if err != nil {
		return dto.InvitationResponse{}, fmt.Errorf("%w: %s", definitions.ErrEventNotFound, err.Error())
	}

	if event.OrganizerID != organizerID {
		return dto.InvitationResponse{}, definitions.ErrNotOrganizer
	}

	// Если активный инвайт уже существует — возвращаем его, не создавая новый
	if existing, err := uc.repo.GetActiveByEvent(ctx, input.EventID.String()); err == nil && existing != nil {
		inviteURL := fmt.Sprintf("%s/invite/%s", uc.baseURL, existing.Token)
		return dto.InvitationResponse{
			InviteURL: inviteURL,
			Token:     existing.Token,
			ExpiresAt: existing.ExpiresAt,
		}, nil
	}

	inv := entity.NewInvitation(input.EventID, organizerID, input.ExpiresIn)
	createdInv, err := uc.repo.Create(ctx, inv)
	if err != nil {
		return dto.InvitationResponse{}, fmt.Errorf("failed to create invitation: %w", err)
	}

	inviteURL := fmt.Sprintf("%s/invite/%s", uc.baseURL, createdInv.Token)

	if uc.log != nil {
		uc.log.Info("invitation generated", slog.String("token", createdInv.Token))
	}

	return dto.InvitationResponse{
		InviteURL: inviteURL,
		Token:     createdInv.Token,
		ExpiresAt: createdInv.ExpiresAt,
	}, nil
}

func (uc *UseCase) SendEmailInvitation(ctx context.Context, input dto.CreateInvitationInput, organizerID uuid.UUID, recipientEmail string) (dto.InvitationResponse, error) {
	resp, err := uc.GenerateInvite(ctx, input, organizerID)
	if err != nil {
		return dto.InvitationResponse{}, err
	}

	if uc.emailService != nil {
		event, err := uc.eventRepo.GetByID(ctx, input.EventID)
		if err != nil {
			return dto.InvitationResponse{}, fmt.Errorf("%w: %s", definitions.ErrEventNotFound, err.Error())
		}
		if err := uc.emailService.SendInvitationEmail(ctx, recipientEmail, event.Title, resp.InviteURL); err != nil {
			if uc.log != nil {
				uc.log.Warn("failed to send invitation email",
					slog.String("email", recipientEmail),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	return resp, nil
}

func (uc *UseCase) JoinByInvite(ctx context.Context, input dto.JoinByInvitationInput) (uuid.UUID, error) {
	if uc.log != nil {
		uc.log.Info("join by invitation started", slog.String("token", input.Token))
	}

	inv, err := uc.repo.GetByToken(ctx, input.Token)
	if err != nil {
		return uuid.Nil, definitions.ErrNotFound
	}

	if !inv.IsValid() {
		return uuid.Nil, definitions.ErrInvalidEventState
	}

	event, err := uc.eventRepo.GetByID(ctx, inv.EventID)
	if err != nil {
		return uuid.Nil, err
	}

	if !event.CanAddParticipants() {
		return uuid.Nil, definitions.ErrInvalidEventState
	}

	_, err = uc.participantUC.Create(ctx, inv.EventID, input.UserID, definitions.ParticipantRoleParticipant)
	if err != nil {
		return uuid.Nil, err
	}

	if uc.log != nil {
		uc.log.Info("user joined via invitation",
			slog.String("event_id", inv.EventID.String()),
			slog.String("user_id", input.UserID.String()),
		)
	}
	return inv.EventID, nil
}
