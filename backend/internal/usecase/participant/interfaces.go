package participant

import (
	"context"

	"secret-santa-backend/internal/entity"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, participant entity.Participant) (entity.Participant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Participant, error)
	GetByEvent(ctx context.Context, eventID uuid.UUID) ([]entity.Participant, error)
	GetByEventPaged(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]entity.Participant, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserAndEvent(ctx context.Context, userID, eventID uuid.UUID) (*entity.Participant, error)
}

type EventRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error)
}
