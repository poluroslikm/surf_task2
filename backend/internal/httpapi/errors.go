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
	CodeBadRequest    = "bad_request"
	CodeUnauthorized  = "unauthorized"
	CodeForbidden     = "forbidden"
	CodeNotFound      = "not_found"
	CodeInternalError = "internal_error"
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
	default:
		writeError(w, http.StatusInternalServerError, CodeInternalError, "Что-то пошло не так. Попробуйте ещё раз позже.")
	}
}
