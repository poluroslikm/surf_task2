package httpapi

import (
	"net/http"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
	"github.com/chef-stol/backend/internal/service"
)

type PushHandler struct {
	svc *service.PushService
}

func NewPushHandler(svc *service.PushService) *PushHandler {
	return &PushHandler{svc: svc}
}

// Register — POST /push/subscriptions (operationId: registerPushSubscription, FEAT-03). Always
// 200 with the stored subscription echoed back — upsert by endpoint, never 409: a
// re-registration of the same endpoint, or the same endpoint changing owner on a shared device,
// are both legitimate, not conflicts.
func (h *PushHandler) Register(w http.ResponseWriter, r *http.Request) {
	client, ok := ClientFromContext(r.Context())
	if !ok {
		mapDomainError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.PushSubscriptionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	sub, err := h.svc.RegisterSubscription(r.Context(), client.ID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.PushSubscriptionResponse{
		Endpoint: sub.Endpoint,
		Keys: dto.PushSubscriptionKeys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	})
}
