package participant

import (
	"context"
	"fmt"

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

func (r *Repository) Create(ctx context.Context, p entity.Participant) (entity.Participant, error) {
	query, args, err := createParticipantQuery().
		Values(p.EventID, p.UserID).
		Suffix("RETURNING id, event_id, user_id, created_at").
		ToSql()

	if err != nil {
		return entity.Participant{}, err
	}

	row := r.db.QueryRow(ctx, query, args...)
	returned, err := scanParticipant(row)
	if err != nil {
		return entity.Participant{}, err
	}

	return *returned, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Participant, error) {
	query := getParticipantByIDQuery(id.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, sql, args...)
	return scanParticipant(row)
}

func (r *Repository) GetByEvent(ctx context.Context, eventID uuid.UUID) ([]entity.Participant, error) {
	query := getParticipantsByEventQuery(eventID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanParticipants(rows)
}

// GetByEventPaged возвращает участников события с пагинацией.
func (r *Repository) GetByEventPaged(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]entity.Participant, int, error) {
	// Запрос с пагинацией
	query := getParticipantsByEventQuery(eventID.String()).
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
	participants, err := scanParticipants(rows)
	if err != nil {
		return nil, 0, err
	}

	// Общее число участников
	countSQL, countArgs, err := countParticipantsByEventQuery(eventID.String()).ToSql()
	if err != nil {
		return participants, 0, nil
	}
	var total int
	_ = r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total)

	return participants, total, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := deleteParticipantQuery(id.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) GetByUserAndEvent(ctx context.Context, userID, eventID uuid.UUID) (*entity.Participant, error) {
	query := getParticipantByUserAndEventQuery(userID.String(), eventID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, sql, args...)
	p, err := scanParticipant(row)
	if err != nil {
		return nil, fmt.Errorf("get participant by user and event: %w", err)
	}
	return p, nil
}
