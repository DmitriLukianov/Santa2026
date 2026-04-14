package verification

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	query := saveCodeQuery().
		Values(email, code, expiresAt)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) GetValidCode(ctx context.Context, email, code string) (bool, error) {
	query := getValidCodeQuery().
		Where("email = ?", email).
		Where("code = ?", code)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	_, err = scanVerificationCode(row)
	if err != nil {
		return false, nil // код не найден или просрочен
	}
	return true, nil
}

func (r *Repository) InvalidateCodes(ctx context.Context, email string) error {
	query := invalidateCodesQuery(email)
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) MarkAsUsed(ctx context.Context, email, code string) error {
	query := markAsUsedQuery().
		Where("email = ?", email).
		Where("code = ?", code)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}
