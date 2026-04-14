package invitation

import (
	"context"

	"secret-santa-backend/internal/entity"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, invitation entity.Invitation) (entity.Invitation, error)
	GetByToken(ctx context.Context, token string) (*entity.Invitation, error)
	GetActiveByEvent(ctx context.Context, eventID string) (*entity.Invitation, error)
}

type EventRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error)
}
