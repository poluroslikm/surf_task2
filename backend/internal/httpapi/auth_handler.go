package httpapi

import (
	"net/http"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
	"github.com/chef-stol/backend/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register — POST /auth/register (operationId: register, FR-1).
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if !validEmail(req.Email) || len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, CodeBadRequest, "Неверные параметры запроса. Проверьте корректность переданных значений.")
		return
	}
	client, token, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAuthResponse(client, token))
}

// Login — POST /auth/login (operationId: login, FR-2).
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	client, token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAuthResponse(client, token))
}

// Logout — POST /auth/logout (operationId: logout).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Logout(r.Context(), bearerToken(r)); err != nil {
		mapDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toAuthResponse(c domain.Client, token string) dto.AuthResponse {
	return dto.AuthResponse{
		Token: token,
		Client: dto.Client{
			ID:        c.ID,
			Email:     c.Email,
			CreatedAt: c.CreatedAt,
		},
	}
}
