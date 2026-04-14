package v1

import (
	"encoding/json"
	"net/http"

	"secret-santa-backend/internal/controller/http/v1/response"
	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/helpers"
	"secret-santa-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ParticipantHandler struct {
	uc      usecase.ParticipantUseCase
	eventUC usecase.EventUseCase
}

func NewParticipantHandler(uc usecase.ParticipantUseCase, eventUC usecase.EventUseCase) *ParticipantHandler {
	return &ParticipantHandler{uc: uc, eventUC: eventUC}
}


func (h *ParticipantHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	participant, err := h.uc.Create(r.Context(), eventID, userID, definitions.ParticipantRoleParticipant)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.ParticipantToResponse(&participant))
}

func (h *ParticipantHandler) GetByEvent(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	requesterID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	participants, err := h.uc.GetByEvent(r.Context(), eventID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	// Только участник или организатор события может видеть список
	isMember := false
	for _, p := range participants {
		if p.UserID == requesterID {
			isMember = true
			break
		}
	}
	if !isMember {
		event, err := h.eventUC.GetByID(r.Context(), eventID)
		if err != nil || event.OrganizerID != requesterID {
			response.WriteHTTPError(w, definitions.ErrForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.ParticipantsToResponse(participants))
}


func (h *ParticipantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	requesterID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	if err := h.uc.Delete(r.Context(), id, requesterID); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
