package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
)

// Machine codes fixed by api/common/models.yaml — do not invent new ones here without
// updating the contract first (see BE-11 in 6-development/BE_IMPLEMENTATION_PLAN.md).
const (
	CodeBadRequest       = "bad_request"
	CodeUnauthorized     = "unauthorized"
	CodeForbidden        = "forbidden"
	CodeNotFound         = "not_found"
	CodeInternalError    = "internal_error"
	CodeSlotFull         = "slot_full"
	CodeSlotCancelled    = "slot_cancelled"
	CodeSlotStarted      = "slot_started"
	CodeAlreadyCancelled = "already_cancelled"
	CodeNotRatable       = "not_ratable"
	CodeAlreadyRated     = "already_rated"
)

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, dto.Error{Code: code, Message: message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// mapDomainError maps sentinel domain errors to the contractual HTTP error response.
// Anything unrecognized falls through to 500 internal_error (common/models.yaml → default).
func mapDomainError(w http.ResponseWriter, err error) {
	// ErrSlotFull carries dynamic data (the real free-seats count for details.available_seats)
	// so it's a type, not a sentinel — checked separately from the errors.Is switch below.
	var slotFull *domain.ErrSlotFull
	if errors.As(err, &slotFull) {
		writeJSON(w, http.StatusConflict, dto.Error{
			Code:    CodeSlotFull,
			Message: "На выбранном слоте не осталось свободных мест.",
			Details: map[string]any{"available_seats": slotFull.AvailableSeats},
		})
		return
	}

	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		// auth/api.yaml fixes this as HTTP 409 with code "bad_request" (not a dedicated
		// "email_taken" code) — matched here exactly as documented, however inconsistent.
		writeError(w, http.StatusConflict, CodeBadRequest, "Этот email уже зарегистрирован.")
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, CodeUnauthorized, "Неверный email или пароль.")
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, CodeUnauthorized, "Требуется авторизация. Передайте действительный токен в заголовке Authorization.")
	case errors.Is(err, domain.ErrSlotNotFound):
		writeError(w, http.StatusNotFound, CodeNotFound, "Запрашиваемый ресурс не найден.")
	case errors.Is(err, domain.ErrInvalidSlotID):
		// createBooking's contract has no documented 404 (unlike getSlot) — an unknown
		// slot_id maps to 400 bad_request instead.
		writeError(w, http.StatusBadRequest, CodeBadRequest, "Неверные параметры запроса. Проверьте корректность переданных значений.")
	case errors.Is(err, domain.ErrSlotCancelled):
		writeError(w, http.StatusGone, CodeSlotCancelled, "Класс отменён студией и более недоступен для записи.")
	case errors.Is(err, domain.ErrInvalidRating):
		writeError(w, http.StatusBadRequest, CodeBadRequest, "Неверные параметры запроса. Проверьте корректность переданных значений.")
	case errors.Is(err, domain.ErrBookingNotFound):
		writeError(w, http.StatusNotFound, CodeNotFound, "Запрашиваемый ресурс не найден.")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, CodeForbidden, "Доступ запрещён. Вы не можете обращаться к данным другого клиента.")
	case errors.Is(err, domain.ErrAlreadyCancelled):
		writeError(w, http.StatusConflict, CodeAlreadyCancelled, "Запись уже отменена.")
	case errors.Is(err, domain.ErrSlotStarted):
		writeError(w, http.StatusUnprocessableEntity, CodeSlotStarted, "Класс уже начался, операция недоступна.")
	case errors.Is(err, domain.ErrAlreadyRated):
		writeError(w, http.StatusConflict, CodeAlreadyRated, "Вы уже оценили этот класс.")
	case errors.Is(err, domain.ErrNotRatable):
		writeError(w, http.StatusUnprocessableEntity, CodeNotRatable, "Оценка доступна только для прошедших неотменённых записей.")
	case errors.Is(err, domain.ErrDuplicateBooking):
		// No dedicated contract code for "already has an active booking on this slot" —
		// mirrors the ErrEmailTaken precedent above (409 + the closest existing code,
		// "bad_request") rather than inventing an undocumented machine code.
		writeError(w, http.StatusConflict, CodeBadRequest, "У вас уже есть активная запись на этот слот.")
	default:
		writeError(w, http.StatusInternalServerError, CodeInternalError, "Что-то пошло не так. Попробуйте ещё раз позже.")
	}
}
