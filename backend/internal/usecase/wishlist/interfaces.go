package wishlist

import (
	"context"

	"secret-santa-backend/internal/entity"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, wishlist entity.Wishlist) (entity.Wishlist, error)
	CreateItem(ctx context.Context, item entity.WishlistItem) (entity.WishlistItem, error)

	GetByParticipant(ctx context.Context, participantID uuid.UUID) (*entity.Wishlist, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wishlist, error)
	GetByID(ctx context.Context, wishlistID uuid.UUID) (*entity.Wishlist, error)
	GetItems(ctx context.Context, wishlistID uuid.UUID) ([]entity.WishlistItem, error)
	GetItemsPaged(ctx context.Context, wishlistID uuid.UUID, limit, offset int) ([]entity.WishlistItem, int, error)
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*entity.WishlistItem, error)
	UpdateItem(ctx context.Context, itemID uuid.UUID, title string, link, imageURL *string, price *float64) error
	DeleteItem(ctx context.Context, itemID uuid.UUID) error
}

type ParticipantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Participant, error)
}

type AssignmentRepository interface {
	GetByEvent(ctx context.Context, eventID uuid.UUID) ([]entity.Assignment, error)
	GetByGiver(ctx context.Context, eventID, giverID uuid.UUID) (*entity.Assignment, error)
}

