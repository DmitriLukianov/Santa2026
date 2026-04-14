package response

import (
	"secret-santa-backend/internal/entity"
	"time"
)

type MessageResponse struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"senderId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func MessageToResponse(m *entity.Message) MessageResponse {
	return MessageResponse{
		ID:        m.ID.String(),
		SenderID:  m.SenderID.String(),
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func MessagesToResponse(messages []entity.Message) []MessageResponse {
	resp := make([]MessageResponse, len(messages))
	for i, m := range messages {
		resp[i] = MessageToResponse(&m)
	}
	return resp
}
