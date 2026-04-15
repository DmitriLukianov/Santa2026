package response

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"secret-santa-backend/internal/definitions"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func WriteHTTPError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	status := http.StatusInternalServerError
	message := "internal server error"

	log.Printf("[HTTP ERROR] %v", err)

	switch {

	case errors.Is(err, definitions.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "Необходима авторизация"

	case errors.Is(err, definitions.ErrNotFound):
		status = http.StatusNotFound
		message = "Объект не найден"

	case errors.Is(err, definitions.ErrEventNotFound):
		status = http.StatusNotFound
		message = "Игра не найдена"

	case errors.Is(err, definitions.ErrWishlistNotFound):
		status = http.StatusNotFound
		message = "Вишлист не найден"

	case errors.Is(err, definitions.ErrParticipantNotFound):
		status = http.StatusNotFound
		message = "Участник не найден"

	case errors.Is(err, definitions.ErrAssignmentNotFound):
		status = http.StatusNotFound
		message = "Результат жеребьёвки не найден"

	case errors.Is(err, definitions.ErrUserNotFound):
		status = http.StatusNotFound
		message = "Пользователь не найден"

	case errors.Is(err, definitions.ErrForbidden):
		status = http.StatusForbidden
		message = "Доступ запрещён"

	case errors.Is(err, definitions.ErrNotOrganizer):
		status = http.StatusForbidden
		message = "Только организатор может выполнить это действие"

	case errors.Is(err, definitions.ErrNotSanta):
		status = http.StatusForbidden
		message = "Вы не являетесь Сантой в этой игре"

	case errors.Is(err, definitions.ErrWishlistVisibilityForbidden):
		status = http.StatusForbidden
		message = "У вас нет доступа к этому вишлисту"

	case errors.Is(err, definitions.ErrConflict):
		status = http.StatusConflict
		message = "Конфликт данных"

	case errors.Is(err, definitions.ErrAlreadyParticipating),
		errors.Is(err, definitions.ErrDuplicateParticipant):
		status = http.StatusConflict
		message = "Вы уже участвуете в этой игре"

	case errors.Is(err, definitions.ErrEventAlreadyFinished):
		status = http.StatusConflict
		message = "Игра уже завершена"

	case errors.Is(err, definitions.ErrInvalidEventState):
		status = http.StatusBadRequest
		message = "Действие недоступно для текущего статуса игры"

	case errors.Is(err, definitions.ErrNotEnoughParticipants):
		status = http.StatusBadRequest
		message = "Недостаточно участников для жеребьёвки"

	case errors.Is(err, definitions.ErrInvalidUserInput):
		status = http.StatusBadRequest
		message = "Некорректные данные"

	case errors.Is(err, definitions.ErrInvalidUUID):
		status = http.StatusBadRequest
		message = "Некорректный идентификатор"

	case errors.Is(err, definitions.ErrInvalidOAuthCode),
		errors.Is(err, definitions.ErrMissingOAuthCode),
		errors.Is(err, definitions.ErrInvalidOAuthUserInfo):
		status = http.StatusBadRequest
		message = "Ошибка авторизации через OAuth"

	case errors.Is(err, definitions.ErrInvalidWishlistVisibility):
		status = http.StatusBadRequest
		message = "Некорректный уровень видимости вишлиста"

	default:
		status = http.StatusInternalServerError
		message = "Внутренняя ошибка сервера"
	}

	writeJSONError(w, status, message)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  status,
	}); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
