package chat

import (
	"context"

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

// CreateMessage — теперь DB-first: возвращает полностью заполненную сущность из БД
func (r *Repository) CreateMessage(ctx context.Context, msg entity.Message) (entity.Message, error) {
	query := createMessageQuery().
		Values(msg.EventID, msg.SenderID, msg.ReceiverID, msg.Content).
		Suffix("RETURNING id, event_id, sender_id, receiver_id, content, created_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.Message{}, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	returned, err := scanMessage(row)
	if err != nil {
		return entity.Message{}, err
	}

	return *returned, nil
}

func (r *Repository) GetMessagesByPair(ctx context.Context, eventID, user1ID, user2ID uuid.UUID) ([]entity.Message, error) {
	return r.GetMessagesByPairPaged(ctx, eventID, user1ID, user2ID, 100, 0)
}

// GetMessagesByPairPaged возвращает сообщения между двумя пользователями с пагинацией.
func (r *Repository) GetMessagesByPairPaged(ctx context.Context, eventID, user1ID, user2ID uuid.UUID, limit, offset int) ([]entity.Message, error) {
	query := getMessagesByPairQuery(eventID.String(), user1ID.String(), user2ID.String()).
		Limit(uint64(limit)).
		Offset(uint64(offset))
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}
