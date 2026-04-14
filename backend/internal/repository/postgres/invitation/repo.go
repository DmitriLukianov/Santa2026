package invitation

import (
	"context"

	"secret-santa-backend/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, i entity.Invitation) (entity.Invitation, error) {
	query := createInvitationQuery().
		Values(i.EventID, i.Token, i.ExpiresAt, i.CreatedBy).
		Suffix("RETURNING id, event_id, token, expires_at, created_by, created_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.Invitation{}, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	returned, err := ScanInvitation(row)
	if err != nil {
		return entity.Invitation{}, err
	}

	return *returned, nil
}

func (r *Repository) GetByToken(ctx context.Context, token string) (*entity.Invitation, error) {
	query, args, err := getInvitationByTokenQuery().
		Where("token = ?", token).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, query, args...)
	return ScanInvitation(row)
}

func (r *Repository) GetActiveByEvent(ctx context.Context, eventID string) (*entity.Invitation, error) {
	query, args, err := getActiveInvitationByEventQuery(eventID).ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, query, args...)
	return ScanInvitation(row)
}
