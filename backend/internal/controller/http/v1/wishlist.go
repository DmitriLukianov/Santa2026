package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"secret-santa-backend/internal/controller/http/v1/request"
	"secret-santa-backend/internal/controller/http/v1/response"
	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/entity"
	"secret-santa-backend/internal/helpers"
	"secret-santa-backend/internal/usecase"
)

type WishlistHandler struct {
	uc            usecase.WishlistUseCase
	participantUC usecase.ParticipantUseCase
}

func NewWishlistHandler(uc usecase.WishlistUseCase, participantUC usecase.ParticipantUseCase) *WishlistHandler {
	return &WishlistHandler{
		uc:            uc,
		participantUC: participantUC,
	}
}

// isWishlistOwner проверяет, что userID является владельцем вишлиста.
// Поддерживает оба типа: вишлист участника события и персональный вишлист.
func (h *WishlistHandler) isWishlistOwner(r *http.Request, wishlist *entity.Wishlist, userID uuid.UUID) bool {
	if wishlist.UserID != nil {
		return *wishlist.UserID == userID
	}
	if wishlist.ParticipantID != nil {
		participant, err := h.participantUC.GetByID(r.Context(), *wishlist.ParticipantID)
		return err == nil && participant.UserID == userID
	}
	return false
}

func (h *WishlistHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	var req request.CreateWishlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}
	if err := helpers.ValidateStruct(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	participant, err := h.participantUC.GetByUserAndEvent(r.Context(), userID, eventID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	wishlist, err := h.uc.Create(r.Context(), participant.ID, req.Visibility)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.WishlistToResponse(&wishlist))
}

func (h *WishlistHandler) GetByParticipant(w http.ResponseWriter, r *http.Request) {
	participantIDStr := chi.URLParam(r, "participantId")
	participantID, err := uuid.Parse(participantIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	eventIDStr := r.URL.Query().Get("eventId")
	if eventIDStr == "" {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	wishlist, err := h.uc.GetForUser(r.Context(), eventID, participantID, userID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.WishlistToResponse(wishlist))
}

// GetByUser возвращает вишлист текущего пользователя.
// Без ?eventId — персональный вишлист (создаётся автоматически если нет).
// С ?eventId — вишлист участника конкретного события.
func (h *WishlistHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	eventIDStr := r.URL.Query().Get("eventId")

	if eventIDStr == "" {
		// Персональный вишлист
		wishlist, err := h.uc.GetOrCreatePersonal(r.Context(), userID)
		if err != nil {
			response.WriteHTTPError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.WishlistToResponse(wishlist))
		return
	}

	// Вишлист участника события
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	participant, err := h.participantUC.GetByUserAndEvent(r.Context(), userID, eventID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	wishlist, err := h.uc.GetByParticipant(r.Context(), participant.ID)
	if err != nil {
		// Вишлист не найден — создаём автоматически
		newWishlist, createErr := h.uc.Create(r.Context(), participant.ID, "santa_only")
		if createErr != nil {
			response.WriteHTTPError(w, createErr)
			return
		}
		wishlist = &newWishlist
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.WishlistToResponse(wishlist))
}

func (h *WishlistHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	wishlistIDStr := chi.URLParam(r, "wishlistId")
	wishlistID, err := uuid.Parse(wishlistIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	wishlist, err := h.uc.GetByID(r.Context(), wishlistID)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrWishlistNotFound)
		return
	}

	if !h.isWishlistOwner(r, wishlist, userID) {
		response.WriteHTTPError(w, definitions.ErrForbidden)
		return
	}

	var req request.CreateWishlistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}
	if err := helpers.ValidateStruct(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	item, err := h.uc.AddItem(r.Context(), wishlistID, req.Title, &req.Link, &req.ImageURL, req.Price)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.WishlistItemToResponse(&item))
}

func (h *WishlistHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	wishlistIDStr := chi.URLParam(r, "wishlistId")
	wishlistID, err := uuid.Parse(wishlistIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	wishlist, err := h.uc.GetByID(r.Context(), wishlistID)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrWishlistNotFound)
		return
	}

	// Проверяем права доступа
	if !h.isWishlistOwner(r, wishlist, userID) {
		// Не владелец — проверяем видимость
		if wishlist.Visibility == definitions.WishlistVisibilitySantaOnly {
			// Для santa_only нужен eventId — Santa должен передать его
			eventIDStr := r.URL.Query().Get("eventId")
			if eventIDStr == "" {
				response.WriteHTTPError(w, definitions.ErrForbidden)
				return
			}
			eventID, err := uuid.Parse(eventIDStr)
			if err != nil {
				response.WriteHTTPError(w, definitions.ErrInvalidUUID)
				return
			}
			// Повторно используем GetForUser: если доступ есть — пропускаем
			if wishlist.ParticipantID != nil {
				if _, err := h.uc.GetForUser(r.Context(), eventID, *wishlist.ParticipantID, userID); err != nil {
					response.WriteHTTPError(w, err)
					return
				}
			} else {
				response.WriteHTTPError(w, definitions.ErrForbidden)
				return
			}
		}
		// public — доступ разрешён всем
	}

	pg := helpers.ParsePagination(r)
	items, total, err := h.uc.GetItemsPaged(r.Context(), wishlistID, pg.Limit, pg.Offset())
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(helpers.NewPagedResponse(response.WishlistItemsToResponse(items), total, pg))
}

func (h *WishlistHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	var req request.UpdateWishlistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}
	if err := helpers.ValidateStruct(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	item, err := h.uc.GetItemByID(r.Context(), itemID)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrWishlistNotFound)
		return
	}

	wishlist, err := h.uc.GetByID(r.Context(), item.WishlistID)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrWishlistNotFound)
		return
	}

	if !h.isWishlistOwner(r, wishlist, userID) {
		response.WriteHTTPError(w, definitions.ErrForbidden)
		return
	}

	updatedItem, err := h.uc.UpdateItem(r.Context(), itemID, req.Title, &req.Link, &req.ImageURL, req.Price)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.WishlistItemToResponse(&updatedItem))
}

func (h *WishlistHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userID, err := helpers.GetUserID(r)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUUID)
		return
	}

	item, err := h.uc.GetItemByID(r.Context(), itemID)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrWishlistNotFound)
		return
	}

	wishlist, err := h.uc.GetByID(r.Context(), item.WishlistID)
	if err != nil {
		response.WriteHTTPError(w, definitions.ErrWishlistNotFound)
		return
	}

	if !h.isWishlistOwner(r, wishlist, userID) {
		response.WriteHTTPError(w, definitions.ErrForbidden)
		return
	}

	if err := h.uc.DeleteItem(r.Context(), itemID); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
