package chat

import (
	"context"
	"fmt"
	"log/slog"

	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/entity"
	"secret-santa-backend/internal/usecase"

	"github.com/google/uuid"
)

type UseCase struct {
	repo            usecase.ChatRepository
	participantRepo usecase.ParticipantRepository
	assignmentRepo  usecase.AssignmentRepository
	log             *slog.Logger
}

func New(repo usecase.ChatRepository, participantRepo usecase.ParticipantRepository, assignmentRepo usecase.AssignmentRepository) *UseCase {
	return &UseCase{
		repo:            repo,
		participantRepo: participantRepo,
		assignmentRepo:  assignmentRepo,
	}
}

func NewWithLogger(repo usecase.ChatRepository, participantRepo usecase.ParticipantRepository, assignmentRepo usecase.AssignmentRepository, log *slog.Logger) *UseCase {
	uc := New(repo, participantRepo, assignmentRepo)
	uc.log = log
	return uc
}

// SendMessageToSanta — отправить сообщение своему Санте (тому, кто дарит мне).
// Использует точечный запрос GetByReceiver — без загрузки всех назначений.
func (uc *UseCase) SendMessageToSanta(ctx context.Context, eventID, userID uuid.UUID, content string) (entity.Message, error) {
	if content == "" {
		return entity.Message{}, definitions.ErrInvalidUserInput
	}
	if len(content) > 2000 {
		return entity.Message{}, definitions.ErrInvalidUserInput
	}

	assignment, err := uc.assignmentRepo.GetByReceiver(ctx, eventID, userID)
	if err != nil {
		return entity.Message{}, definitions.ErrNotSanta
	}

	msg := entity.NewMessage(eventID, userID, assignment.GiverID, content)
	return uc.repo.CreateMessage(ctx, msg)
}

// SendMessage — Санта отправляет сообщение своему получателю.
// Использует точечный запрос GetByGiver — без загрузки всех назначений.
func (uc *UseCase) SendMessage(ctx context.Context, eventID, userID uuid.UUID, content string) (entity.Message, error) {
	if content == "" {
		return entity.Message{}, definitions.ErrInvalidUserInput
	}
	if len(content) > 2000 {
		return entity.Message{}, definitions.ErrInvalidUserInput
	}

	if uc.log != nil {
		uc.log.Info("send message started",
			slog.String("event_id", eventID.String()),
			slog.String("user_id", userID.String()),
		)
	}

	assignment, err := uc.assignmentRepo.GetByGiver(ctx, eventID, userID)
	if err != nil {
		return entity.Message{}, definitions.ErrNotSanta
	}

	msg := entity.NewMessage(eventID, userID, assignment.ReceiverID, content)
	createdMsg, err := uc.repo.CreateMessage(ctx, msg)
	if err != nil {
		return entity.Message{}, fmt.Errorf("failed to create message: %w", err)
	}

	if uc.log != nil {
		uc.log.Info("message sent successfully",
			slog.String("message_id", createdMsg.ID.String()),
		)
	}
	return createdMsg, nil
}

// GetSenderChat — сообщения от моего Санты (где я получатель).
func (uc *UseCase) GetSenderChat(ctx context.Context, eventID, userID uuid.UUID) ([]entity.Message, error) {
	assignment, err := uc.assignmentRepo.GetByReceiver(ctx, eventID, userID)
	if err != nil {
		// Назначения нет — чат пуст, не ошибка
		return []entity.Message{}, nil
	}
	return uc.repo.GetMessagesByPair(ctx, eventID, assignment.GiverID, userID)
}

// GetRecipientChat — сообщения Санты своему получателю.
func (uc *UseCase) GetRecipientChat(ctx context.Context, eventID, userID uuid.UUID) ([]entity.Message, error) {
	if uc.log != nil {
		uc.log.Info("get recipient chat started",
			slog.String("event_id", eventID.String()),
			slog.String("user_id", userID.String()),
		)
	}

	assignment, err := uc.assignmentRepo.GetByGiver(ctx, eventID, userID)
	if err != nil {
		return nil, definitions.ErrNotSanta
	}
	return uc.repo.GetMessagesByPair(ctx, eventID, userID, assignment.ReceiverID)
}

