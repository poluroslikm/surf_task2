package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
	"github.com/chef-stol/backend/internal/service"
)

type SlotsHandler struct {
	svc *service.SlotsService
}

func NewSlotsHandler(svc *service.SlotsService) *SlotsHandler {
	return &SlotsHandler{svc: svc}
}

// List — GET /slots (operationId: listSlots, FR-3/FR-4/R-027).
func (h *SlotsHandler) List(w http.ResponseWriter, r *http.Request) {
	from, ok := parseOptionalTime(w, r, "date_from")
	if !ok {
		return
	}
	to, ok := parseOptionalTime(w, r, "date_to")
	if !ok {
		return
	}
	slots, err := h.svc.List(r.Context(), from, to)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.SlotListResponse{Items: toSlotDTOs(slots)})
}

// Get — GET /slots/{slotId} (operationId: getSlot).
func (h *SlotsHandler) Get(w http.ResponseWriter, r *http.Request) {
	slot, err := h.svc.Get(r.Context(), chi.URLParam(r, "slotId"))
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSlotDTO(slot))
}

func parseOptionalTime(w http.ResponseWriter, r *http.Request, param string) (*time.Time, bool) {
	raw := r.URL.Query().Get(param)
	if raw == "" {
		return nil, true
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, "Неверные параметры запроса. Проверьте корректность переданных значений.")
		return nil, false
	}
	return &t, true
}

func toSlotDTO(s domain.Slot) dto.Slot {
	return dto.Slot{
		ID: s.ID,
		Program: dto.Program{
			ID:          s.Program.ID,
			Name:        s.Program.Name,
			Difficulty:  string(s.Program.Difficulty),
			PhotoURL:    s.Program.PhotoURL,
			Ingredients: s.Program.Ingredients,
			Allergens:   s.Program.Allergens,
		},
		Chef: dto.Chef{
			ID:   s.Chef.ID,
			Name: s.Chef.Name,
		},
		StartAt:            s.StartAt,
		TotalSeats:         s.TotalSeats,
		FreeSeats:          s.FreeSeats,
		Price:              s.Price,
		Status:             string(s.Status),
		CancellationReason: s.CancellationReason,
	}
}

func toSlotDTOs(slots []domain.Slot) []dto.Slot {
	out := make([]dto.Slot, 0, len(slots))
	for _, s := range slots {
		out = append(out, toSlotDTO(s))
	}
	return out
}
