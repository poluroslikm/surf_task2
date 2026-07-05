package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/service"
)

// InternalHandler backs dev-only tooling endpoints that are deliberately NOT part of api/ (see
// backend/README.md "Правила декомпозиции") — not client-facing, so responses here are plain
// text/JSON, not required to match api/common/models.yaml's Error schema.
type InternalHandler struct {
	svc *service.InternalService
}

func NewInternalHandler(svc *service.InternalService) *InternalHandler {
	return &InternalHandler{svc: svc}
}

type forceCancelSlotRequest struct {
	Reason string `json:"reason"`
}

// ForceCancelSlot — POST /internal/slots/{slotId}/force-cancel (FEAT-05, dev-only, guarded by
// RequireInternalToken, not RequireAuth — this isn't a client session). Simulates FR-17 (studio
// cancels a slot): cascades to FR-18 (re-booking blocked, slot no longer scheduled) and FR-22
// (push to every affected client).
func (h *InternalHandler) ForceCancelSlot(w http.ResponseWriter, r *http.Request) {
	var req forceCancelSlotRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Reason == "" {
		http.Error(w, "reason is required", http.StatusBadRequest)
		return
	}

	slotID := chi.URLParam(r, "slotId")
	if err := h.svc.ForceCancelSlot(r.Context(), slotID, req.Reason); err != nil {
		switch {
		case errors.Is(err, domain.ErrSlotNotFound):
			http.Error(w, "slot not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrSlotAlreadyCancelled):
			http.Error(w, "slot already cancelled", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
