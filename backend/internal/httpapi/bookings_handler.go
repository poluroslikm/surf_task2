package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
	"github.com/chef-stol/backend/internal/service"
)

type BookingsHandler struct {
	svc *service.BookingsService
}

func NewBookingsHandler(svc *service.BookingsService) *BookingsHandler {
	return &BookingsHandler{svc: svc}
}

// Create — POST /bookings (operationId: createBooking, FR-7, NFR-6). Deliberate scope
// decision (see task/report): the contract's createBooking has no documented 422
// slot_started, so a slot whose start_at is already in the past is still bookable as long as
// it's scheduled with free_seats > 0 — no undocumented stricter check is added here.
func (h *BookingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	client, ok := ClientFromContext(r.Context())
	if !ok {
		mapDomainError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.CreateBookingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	equipment := domain.EquipmentChoice(req.EquipmentChoice)
	if equipment != domain.EquipmentOwn && equipment != domain.EquipmentRental {
		writeError(w, http.StatusBadRequest, CodeBadRequest, "Неверные параметры запроса. Проверьте корректность переданных значений.")
		return
	}
	booking, err := h.svc.Create(r.Context(), client.ID, req.SlotID, equipment)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toBookingDTO(booking))
}

// List — GET /bookings (operationId: listBookings). Scoped to the authenticated client only
// (NFR-8); no pagination or status filter, matching api/bookings/api.yaml.
func (h *BookingsHandler) List(w http.ResponseWriter, r *http.Request) {
	client, ok := ClientFromContext(r.Context())
	if !ok {
		mapDomainError(w, domain.ErrUnauthorized)
		return
	}
	bookings, err := h.svc.List(r.Context(), client.ID)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.BookingListResponse{Items: toBookingSummaryDTOs(bookings)})
}

// Get — GET /bookings/{bookingId} (operationId: getBooking).
func (h *BookingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	client, ok := ClientFromContext(r.Context())
	if !ok {
		mapDomainError(w, domain.ErrUnauthorized)
		return
	}
	booking, err := h.svc.Get(r.Context(), client.ID, chi.URLParam(r, "bookingId"))
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBookingDTO(booking))
}

// Cancel — POST /bookings/{bookingId}/cancel (operationId: cancelBooking, FR-15/FR-16).
func (h *BookingsHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	client, ok := ClientFromContext(r.Context())
	if !ok {
		mapDomainError(w, domain.ErrUnauthorized)
		return
	}
	booking, err := h.svc.Cancel(r.Context(), client.ID, chi.URLParam(r, "bookingId"))
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBookingDTO(booking))
}

// SubmitRating — POST /bookings/{bookingId}/rating (operationId: submitRating, FR-19/FR-20).
func (h *BookingsHandler) SubmitRating(w http.ResponseWriter, r *http.Request) {
	client, ok := ClientFromContext(r.Context())
	if !ok {
		mapDomainError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.SubmitRatingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	booking, err := h.svc.SubmitRating(r.Context(), client.ID, chi.URLParam(r, "bookingId"), req.Stars, req.Comment)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBookingDTO(booking))
}

func toBookingDTO(b domain.Booking) dto.Booking {
	out := dto.Booking{
		ID:                    b.ID,
		SlotID:                b.SlotID,
		ClientID:              b.ClientID,
		EquipmentChoice:       string(b.EquipmentChoice),
		Status:                string(b.Status),
		PriceTotal:            b.PriceTotal,
		FreeCancellationUntil: b.FreeCancellationUntil,
		CreatedAt:             b.CreatedAt,
		CancelledAt:           b.CancelledAt,
		Slot:                  toSlotDTO(b.Slot),
	}
	if b.Rating != nil {
		out.Rating = &dto.Rating{
			Stars:     b.Rating.Stars,
			Comment:   b.Rating.Comment,
			CreatedAt: b.Rating.CreatedAt,
		}
	}
	return out
}

func toBookingSummaryDTO(b domain.Booking) dto.BookingSummary {
	return dto.BookingSummary{
		ID:              b.ID,
		SlotID:          b.SlotID,
		EquipmentChoice: string(b.EquipmentChoice),
		Status:          string(b.Status),
		PriceTotal:      b.PriceTotal,
		CreatedAt:       b.CreatedAt,
		CancelledAt:     b.CancelledAt,
		Slot:            toSlotDTO(b.Slot),
	}
}

func toBookingSummaryDTOs(bookings []domain.Booking) []dto.BookingSummary {
	out := make([]dto.BookingSummary, 0, len(bookings))
	for _, b := range bookings {
		out = append(out, toBookingSummaryDTO(b))
	}
	return out
}
