package assignment

import (
	"context"
	"fmt"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create — теперь DB-first: возвращает полностью заполненную сущность из БД
func (r *Repository) Create(ctx context.Context, a entity.Assignment) (entity.Assignment, error) {
	query := createAssignmentQuery().
		Values(a.EventID, a.GiverID, a.ReceiverID).
		Suffix("RETURNING id, event_id, giver_id, receiver_id, created_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.Assignment{}, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	returned, err := scanAssignment(row)
	if err != nil {
		return entity.Assignment{}, err
	}

	return *returned, nil
}

func (r *Repository) GetByEvent(ctx context.Context, eventID uuid.UUID) ([]entity.Assignment, error) {
	query := getAssignmentsByEventQuery(eventID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAssignments(rows)
}

// GetByGiver возвращает назначение конкретного дарителя без загрузки всего списка.
func (r *Repository) GetByGiver(ctx context.Context, eventID, giverID uuid.UUID) (*entity.Assignment, error) {
	query := getAssignmentByGiverQuery(eventID.String(), giverID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, sql, args...)
	a, err := scanAssignmentWithName(row)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// GetByReceiver возвращает назначение для получателя (кто его Санта).
func (r *Repository) GetByReceiver(ctx context.Context, eventID, receiverID uuid.UUID) (*entity.Assignment, error) {
	query := getAssignmentByReceiverQuery(eventID.String(), receiverID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, sql, args...)
	a, err := scanAssignmentWithName(row)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repository) DeleteByEvent(ctx context.Context, eventID uuid.UUID) error {
	query := deleteAssignmentsByEventQuery(eventID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) TransactionalDraw(ctx context.Context, eventID uuid.UUID, assignments []entity.Assignment, newStatus definitions.EventStatus) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.deleteByEventTx(ctx, tx, eventID); err != nil {
		return fmt.Errorf("failed to delete old assignments: %w", err)
	}

	for _, a := range assignments {
		if err := r.createTx(ctx, tx, a); err != nil {
			return fmt.Errorf("failed to create assignment: %w", err)
		}
	}

	if err := r.updateEventStatusTx(ctx, tx, eventID, newStatus); err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) deleteByEventTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID) error {
	query := deleteAssignmentsByEventQuery(eventID.String())
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) createTx(ctx context.Context, tx pgx.Tx, a entity.Assignment) error {
	query := createAssignmentQuery().
		Values(a.EventID, a.GiverID, a.ReceiverID)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	return err
}

func (r *Repository) updateEventStatusTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, status definitions.EventStatus) error {
	_, err := tx.Exec(ctx, `
		UPDATE events 
		SET status = $1, updated_at = now() 
		WHERE id = $2`,
		status, eventID)
	return err
}
