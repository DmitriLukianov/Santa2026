package wishlist

import (
	"context"
	"fmt"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, w entity.Wishlist) (entity.Wishlist, error) {
	var sql string
	var args []interface{}
	var err error

	if w.ParticipantID != nil {
		// Вишлист участника события
		existing, err := r.GetByParticipant(ctx, *w.ParticipantID)
		if err == nil && existing != nil {
			return entity.Wishlist{}, fmt.Errorf("wishlist already exists for participant %s: %w", w.ParticipantID, definitions.ErrConflict)
		}
		sql, args, err = createParticipantWishlistQuery(*w.ParticipantID, w.Visibility).ToSql()
	} else if w.UserID != nil {
		// Персональный вишлист пользователя
		existing, err := r.GetByUserID(ctx, *w.UserID)
		if err == nil && existing != nil {
			return entity.Wishlist{}, fmt.Errorf("personal wishlist already exists: %w", definitions.ErrConflict)
		}
		sql, args, err = createPersonalWishlistQuery(*w.UserID).ToSql()
	} else {
		return entity.Wishlist{}, definitions.ErrInvalidUserInput
	}

	if err != nil {
		return entity.Wishlist{}, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	returned, err := scanWishlist(row)
	if err != nil {
		return entity.Wishlist{}, err
	}
	return *returned, nil
}

func (r *Repository) CreateItem(ctx context.Context, item entity.WishlistItem) (entity.WishlistItem, error) {
	query, args, err := createWishlistItemQuery().
		Values(item.WishlistID, item.Title, item.Link, item.ImageURL, item.Price).
		Suffix("RETURNING id, wishlist_id, title, link, image_url, price, created_at").
		ToSql()

	if err != nil {
		return entity.WishlistItem{}, err
	}

	row := r.db.QueryRow(ctx, query, args...)
	returned, err := scanWishlistItem(row)
	if err != nil {
		return entity.WishlistItem{}, err
	}

	return *returned, nil
}

func (r *Repository) GetByParticipant(ctx context.Context, participantID uuid.UUID) (*entity.Wishlist, error) {
	query := getWishlistByParticipantQuery(participantID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, sql, args...)
	return scanWishlist(row)
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wishlist, error) {
	query := getWishlistByUserIDQuery(userID)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, sql, args...)
	return scanWishlist(row)
}

func (r *Repository) GetItems(ctx context.Context, wishlistID uuid.UUID) ([]entity.WishlistItem, error) {
	items, _, err := r.GetItemsPaged(ctx, wishlistID, 200, 0)
	return items, err
}

// GetItemsPaged возвращает элементы вишлиста с пагинацией.
func (r *Repository) GetItemsPaged(ctx context.Context, wishlistID uuid.UUID, limit, offset int) ([]entity.WishlistItem, int, error) {
	query := getWishlistItemsQuery(wishlistID.String()).
		Limit(uint64(limit)).
		Offset(uint64(offset))
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanWishlistItems(rows)
	if err != nil {
		return nil, 0, err
	}

	countSQL, countArgs, err := countWishlistItemsQuery(wishlistID.String()).ToSql()
	if err != nil {
		return items, 0, nil
	}
	var total int
	_ = r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total)

	return items, total, nil
}

func (r *Repository) UpdateItem(ctx context.Context, itemID uuid.UUID, title string, link, imageURL *string, price *float64) error {
	query := updateWishlistItemQuery(itemID.String()).
		Set("title", title).
		Set("link", link).
		Set("image_url", imageURL).
		Set("price", price)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	query := deleteWishlistItemQuery(itemID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) GetItemByID(ctx context.Context, itemID uuid.UUID) (*entity.WishlistItem, error) {
	if itemID == uuid.Nil {
		return nil, definitions.ErrInvalidUserInput
	}

	query := getWishlistItemByIDQuery(itemID)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return scanWishlistItem(row)
}

func (r *Repository) GetByID(ctx context.Context, wishlistID uuid.UUID) (*entity.Wishlist, error) {
	if wishlistID == uuid.Nil {
		return nil, definitions.ErrInvalidUserInput
	}

	query := getWishlistByIDQuery(wishlistID)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return scanWishlist(row)
}
