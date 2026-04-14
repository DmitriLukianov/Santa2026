package event

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/dto"
	"secret-santa-backend/internal/entity"
	participant "secret-santa-backend/internal/usecase/participant"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/google/uuid"
)

type UseCase struct {
	repo            Repository
	participantRepo participant.Repository
	log             *slog.Logger
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func NewWithLogger(repo Repository, participantRepo participant.Repository, log *slog.Logger) *UseCase {
	return &UseCase{
		repo:            repo,
		participantRepo: participantRepo,
		log:             log,
	}
}

func (uc *UseCase) Create(ctx context.Context, input dto.CreateEventInput, organizerID uuid.UUID) (entity.Event, error) {
	if uc.log != nil {
		uc.log.Info("create event started",
			slog.String("organizer_id", organizerID.String()),
			slog.String("title", input.Title),
		)
	}

	if organizerID == uuid.Nil {
		return entity.Event{}, definitions.ErrInvalidUserInput
	}

	startDate := time.Now()
	if input.StartDate != nil {
		startDate = *input.StartDate
	}

	event := entity.NewEvent(
		input.Title,
		organizerID,
		input.OrganizerNotes,
		startDate,
		input.DrawDate,
	)

	createdEvent, err := uc.repo.Create(ctx, event)
	if err != nil {
		if uc.log != nil {
			uc.log.Error("failed to create event", slog.String("error", err.Error()))
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.Event{}, definitions.ErrConflict
		}
		return entity.Event{}, fmt.Errorf("failed to create event: %w", err)
	}

	if input.WantParticipate {
		organizerParticipant := entity.NewParticipant(createdEvent.ID, organizerID)
		if _, err = uc.participantRepo.Create(ctx, organizerParticipant); err != nil {
			if uc.log != nil {
				uc.log.Error("failed to create organizer participant", slog.String("error", err.Error()))
			}
			return entity.Event{}, fmt.Errorf("failed to create organizer participant: %w", err)
		}
	}

	if uc.log != nil {
		uc.log.Info("event created successfully",
			slog.String("event_id", createdEvent.ID.String()),
		)
	}
	return createdEvent, nil
}

func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	if id == uuid.Nil {
		return nil, definitions.ErrInvalidUserInput
	}
	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, definitions.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return event, nil
}

func (uc *UseCase) Update(ctx context.Context, id, userID uuid.UUID, input dto.UpdateEventInput) error {
	if id == uuid.Nil || userID == uuid.Nil {
		return definitions.ErrInvalidUserInput
	}

	eventPtr, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return definitions.ErrEventNotFound
	}

	if eventPtr.OrganizerID != userID {
		return definitions.ErrNotOrganizer
	}

	if uc.log != nil {
		uc.log.Info("update event started", slog.String("event_id", id.String()))
	}

	if err := uc.repo.Update(ctx, id, input); err != nil {
		if uc.log != nil {
			uc.log.Error("failed to update event", slog.String("error", err.Error()))
		}
		return err
	}

	if uc.log != nil {
		uc.log.Info("event updated successfully", slog.String("event_id", id.String()))
	}
	return nil
}



func (uc *UseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil || userID == uuid.Nil {
		return definitions.ErrInvalidUserInput
	}

	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return definitions.ErrEventNotFound
	}

	if event.OrganizerID != userID {
		return definitions.ErrNotOrganizer
	}

	if uc.log != nil {
		uc.log.Info("delete event", slog.String("event_id", id.String()))
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *UseCase) Activate(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil || userID == uuid.Nil {
		return definitions.ErrInvalidUserInput
	}

	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return definitions.ErrEventNotFound
	}

	if event.OrganizerID != userID {
		return definitions.ErrNotOrganizer
	}

	if err := event.TransitionTo(definitions.EventStatusGifting); err != nil {
		return err
	}

	return uc.repo.UpdateStatus(ctx, id, definitions.EventStatusGifting)
}

func (uc *UseCase) Finish(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil || userID == uuid.Nil {
		return definitions.ErrInvalidUserInput
	}

	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return definitions.ErrEventNotFound
	}

	if event.OrganizerID != userID {
		return definitions.ErrNotOrganizer
	}

	if err := event.TransitionTo(definitions.EventStatusFinished); err != nil {
		return err
	}

	return uc.repo.UpdateStatus(ctx, id, definitions.EventStatusFinished)
}

func (uc *UseCase) GetMyEvents(ctx context.Context, userID uuid.UUID) ([]entity.Event, error) {
	if userID == uuid.Nil {
		return nil, definitions.ErrInvalidUserInput
	}

	events, err := uc.repo.GetEventsForUser(ctx, userID)
	if err != nil {
		if uc.log != nil {
			uc.log.Error("failed to get events for user",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
		}
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	if events == nil {
		return []entity.Event{}, nil
	}
	return events, nil
}

func (uc *UseCase) UpdateStatus(ctx context.Context, id uuid.UUID, status definitions.EventStatus) error {
	if id == uuid.Nil {
		return definitions.ErrInvalidUserInput
	}
	return uc.repo.UpdateStatus(ctx, id, status)
}
