package helpers

import (
	"net/http"

	"secret-santa-backend/internal/definitions"

	"github.com/google/uuid"
)

func GetUserID(r *http.Request) (uuid.UUID, error) {
	val := r.Context().Value(definitions.UserIDKey)
	if val == nil {
		return uuid.Nil, definitions.ErrUnauthorized
	}

	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, definitions.ErrUnauthorized
	}

	return id, nil
}
