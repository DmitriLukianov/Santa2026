package event

import (
	"context"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/dto"
	"secret-santa-backend/internal/entity"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, e entity.Event) (entity.Event, error) {
	query, args, err := createEventQuery().
		Values(
			e.Title,
			e.OrganizerNotes,
			e.OrganizerID,
			e.StartDate,
			e.DrawDate,
			e.Status,
		).
		Suffix("RETURNING id, title, organizer_notes, organizer_id, " +
			"start_date, draw_date, status, created_at, updated_at").
		ToSql()

	if err != nil {
		return entity.Event{}, err
	}

	row := r.db.QueryRow(ctx, query, args...)
	returnedEvent, err := ScanEvent(row)
	if err != nil {
		return entity.Event{}, err
	}

	return *returnedEvent, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	query, args, err := getEventQuery().
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, query, args...)
	return ScanEvent(row)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input dto.UpdateEventInput) error {
	q := updateEventQuery().Set("updated_at", squirrel.Expr("NOW()"))

	if input.Title != nil {
		q = q.Set("title", *input.Title)
	}
	if input.OrganizerNotes != nil {
		q = q.Set("organizer_notes", *input.OrganizerNotes)
	}
	if input.StartDate != nil {
		q = q.Set("start_date", *input.StartDate)
	}
	if input.DrawDate != nil {
		q = q.Set("draw_date", *input.DrawDate)
	}

	query, args, err := q.Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status definitions.EventStatus) error {
	query, args, err := updateEventQuery().
		Set("status", status).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete: ставим метку deleted_at вместо физического удаления.
	// Данные (участники, жеребьёвка, чат) сохраняются для истории.
	query, args, err := updateEventQuery().
		Set("deleted_at", squirrel.Expr("NOW()")).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) GetDueForDraw(ctx context.Context) ([]entity.Event, error) {
	query, args, err := getEventQuery().
		Where("draw_date <= NOW() AND status = ?", definitions.EventStatusRegistration).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ScanEvents(rows)
}

func (r *Repository) GetEventsForUser(ctx context.Context, userID uuid.UUID) ([]entity.Event, error) {
	query, args, err := getEventsForUserQuery().
		Where("events.organizer_id = ? OR p.user_id = ?", userID, userID).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ScanEvents(rows)
}
