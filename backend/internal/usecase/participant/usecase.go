package participant

import (
	"context"
	"log/slog"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/entity"
	"secret-santa-backend/internal/helpers"

	"github.com/google/uuid"
)

type UseCase struct {
	repo      Repository
	eventRepo EventRepository
	log       *slog.Logger
}

func New(repo Repository, eventRepo EventRepository) *UseCase {
	return &UseCase{repo: repo, eventRepo: eventRepo}
}

func NewWithLogger(repo Repository, eventRepo EventRepository, log *slog.Logger) *UseCase {
	return &UseCase{repo: repo, eventRepo: eventRepo, log: log}
}

func (uc *UseCase) Create(ctx context.Context, eventID, userID uuid.UUID, role string) (entity.Participant, error) {
	if uc.log != nil {
		uc.log.Info("create participant started",
			slog.String("event_id", eventID.String()),
			slog.String("user_id", userID.String()),
		)
	}

	if eventID == uuid.Nil || userID == uuid.Nil {
		return entity.Participant{}, definitions.ErrInvalidUserInput
	}

	event, err := uc.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return entity.Participant{}, definitions.ErrEventNotFound
	}
	if !event.CanAddParticipants() {
		return entity.Participant{}, definitions.ErrInvalidEventState
	}

	participant := entity.NewParticipant(eventID, userID)
	created, err := uc.repo.Create(ctx, participant)
	if err != nil {
		if uc.log != nil {
			uc.log.Error("failed to create participant", slog.String("error", err.Error()))
		}
		if helpers.IsDuplicateError(err) {
			return entity.Participant{}, definitions.ErrAlreadyParticipating
		}
		return entity.Participant{}, err
	}

	if uc.log != nil {
		uc.log.Info("participant created successfully",
			slog.String("participant_id", created.ID.String()),
		)
	}
	return created, nil
}

func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Participant, error) {
	if id == uuid.Nil {
		return nil, definitions.ErrInvalidUserInput
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *UseCase) GetByEvent(ctx context.Context, eventID uuid.UUID) ([]entity.Participant, error) {
	if eventID == uuid.Nil {
		return nil, definitions.ErrInvalidUserInput
	}
	return uc.repo.GetByEvent(ctx, eventID)
}

func (uc *UseCase) GetByEventPaged(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]entity.Participant, int, error) {
	if eventID == uuid.Nil {
		return nil, 0, definitions.ErrInvalidUserInput
	}
	return uc.repo.GetByEventPaged(ctx, eventID, limit, offset)
}

func (uc *UseCase) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if id == uuid.Nil || requesterID == uuid.Nil {
		return definitions.ErrInvalidUserInput
	}

	if uc.log != nil {
		uc.log.Info("delete participant started",
			slog.String("participant_id", id.String()),
			slog.String("requester_id", requesterID.String()),
		)
	}

	p, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return definitions.ErrParticipantNotFound
	}

	// Участник может удалить себя сам; организатор события — любого участника
	event, err := uc.eventRepo.GetByID(ctx, p.EventID)
	if err != nil {
		return definitions.ErrEventNotFound
	}

	if p.UserID != requesterID && event.OrganizerID != requesterID {
		return definitions.ErrForbidden
	}

	// Нельзя покидать/удалять участника после начала жеребьёвки
	if event.Status == definitions.EventStatusGifting || event.Status == definitions.EventStatusFinished {
		return definitions.ErrInvalidEventState
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		if uc.log != nil {
			uc.log.Error("failed to delete participant", slog.String("error", err.Error()))
		}
		return err
	}

	if uc.log != nil {
		uc.log.Info("participant deleted successfully", slog.String("participant_id", id.String()))
	}
	return nil
}

func (uc *UseCase) GetByUserAndEvent(ctx context.Context, userID, eventID uuid.UUID) (*entity.Participant, error) {
	if userID == uuid.Nil || eventID == uuid.Nil {
		return nil, definitions.ErrInvalidUserInput
	}
	return uc.repo.GetByUserAndEvent(ctx, userID, eventID)
}
