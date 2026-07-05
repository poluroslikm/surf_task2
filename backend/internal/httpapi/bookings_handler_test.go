package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
)

// doAuthedJSON mirrors doJSON (auth_handler_test.go) but also sets an Authorization: Bearer
// header when token is non-empty, since every bookings endpoint is mounted behind RequireAuth.
func doAuthedJSON(t *testing.T, router http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// mustRegister registers a fresh client through the real /auth/register endpoint (so the
// session token is issued by the real AuthService, not hand-crafted) and returns the decoded
// response.
func mustRegister(t *testing.T, router http.Handler, email string) dto.AuthResponse {
	t.Helper()
	resp := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    email,
		Password: "password1",
	})
	if resp.Code != http.StatusCreated {
		t.Fatalf("register(%s) status = %d, want 201, body: %s", email, resp.Code, resp.Body.String())
	}
	var auth dto.AuthResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &auth); err != nil {
		t.Fatalf("decode register(%s) response: %v", email, err)
	}
	return auth
}

func sampleSlot(id string) domain.Slot {
	return domain.Slot{
		ID: id,
		Program: domain.Program{
			ID:          "prog-1",
			Name:        "Паста ручной работы",
			Difficulty:  domain.DifficultyNovice,
			PhotoURL:    "https://example.com/photo.jpg",
			Ingredients: []string{"мука", "яйца"},
			Allergens:   []string{"глютен"},
		},
		Chef:       domain.Chef{ID: "chef-1", Name: "Иван Иванов"},
		StartAt:    time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC),
		TotalSeats: 12,
		FreeSeats:  10,
		Price:      2500,
		Status:     domain.SlotScheduled,
	}
}

func sampleBooking(id, clientID string, status domain.BookingStatus) domain.Booking {
	return domain.Booking{
		ID:                    id,
		SlotID:                "slot-1",
		ClientID:              clientID,
		EquipmentChoice:       domain.EquipmentOwn,
		Status:                status,
		PriceTotal:            2500,
		FreeCancellationUntil: time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC),
		CreatedAt:             time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
		Slot:                  sampleSlot("slot-1"),
	}
}

// ---- createBooking ----

func TestBookingsHandler_CreateSuccess(t *testing.T) {
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(newFakeAuthRepo(), &fakeSlotsRepo{}, bookingsRepo)
	auth := mustRegister(t, router, "create-ok@example.com")
	bookingsRepo.createResult = sampleBooking("booking-1", auth.Client.ID, domain.BookingActive)

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings", auth.Token, dto.CreateBookingRequest{
		SlotID:          "slot-1",
		EquipmentChoice: "own",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", w.Code, w.Body.String())
	}
	var got dto.Booking
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "booking-1" || got.Status != "active" || got.PriceTotal != 2500 {
		t.Fatalf("booking mismatch: %+v", got)
	}
	if got.Slot.ID != "slot-1" || got.Slot.Chef.Name != "Иван Иванов" {
		t.Fatalf("nested slot mismatch: %+v", got.Slot)
	}
}

func TestBookingsHandler_CreateInvalidEquipmentChoiceIs400(t *testing.T) {
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(newFakeAuthRepo(), &fakeSlotsRepo{}, bookingsRepo)
	auth := mustRegister(t, router, "create-badequip@example.com")

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings", auth.Token, dto.CreateBookingRequest{
		SlotID:          "slot-1",
		EquipmentChoice: "not-a-real-choice",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeBadRequest {
		t.Fatalf("code = %q, want %q", body.Code, CodeBadRequest)
	}
}

func TestBookingsHandler_CreateSlotFullIs409WithAvailableSeats(t *testing.T) {
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(newFakeAuthRepo(), &fakeSlotsRepo{}, bookingsRepo)
	auth := mustRegister(t, router, "create-full@example.com")
	bookingsRepo.createErr = &domain.ErrSlotFull{AvailableSeats: 0}

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings", auth.Token, dto.CreateBookingRequest{
		SlotID:          "slot-1",
		EquipmentChoice: "own",
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeSlotFull {
		t.Fatalf("code = %q, want %q", body.Code, CodeSlotFull)
	}
	seats, ok := body.Details["available_seats"]
	if !ok {
		t.Fatalf("details missing available_seats: %+v", body.Details)
	}
	if seats != float64(0) {
		t.Fatalf("details.available_seats = %v, want 0", seats)
	}
}

func TestBookingsHandler_CreateSlotCancelledIs410(t *testing.T) {
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(newFakeAuthRepo(), &fakeSlotsRepo{}, bookingsRepo)
	auth := mustRegister(t, router, "create-cancelled@example.com")
	bookingsRepo.createErr = domain.ErrSlotCancelled

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings", auth.Token, dto.CreateBookingRequest{
		SlotID:          "slot-1",
		EquipmentChoice: "own",
	})

	if w.Code != http.StatusGone {
		t.Fatalf("status = %d, want 410, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeSlotCancelled {
		t.Fatalf("code = %q, want %q", body.Code, CodeSlotCancelled)
	}
}

func TestBookingsHandler_CreateNoAuthHeaderIs401(t *testing.T) {
	router := newTestRouterWithBookings(newFakeAuthRepo(), &fakeSlotsRepo{}, newFakeBookingsRepo())

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings", "", dto.CreateBookingRequest{
		SlotID:          "slot-1",
		EquipmentChoice: "own",
	})

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body: %s", w.Code, w.Body.String())
	}
}

// ---- listBookings ----

func TestBookingsHandler_ListScopedToAuthenticatedClient(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)

	client1 := mustRegister(t, router, "list-client1@example.com")
	client2 := mustRegister(t, router, "list-client2@example.com")
	bookingsRepo.bookings["booking-mine-1"] = sampleBooking("booking-mine-1", client1.Client.ID, domain.BookingActive)
	bookingsRepo.bookings["booking-mine-2"] = sampleBooking("booking-mine-2", client1.Client.ID, domain.BookingCancelled)
	bookingsRepo.bookings["booking-other"] = sampleBooking("booking-other", client2.Client.ID, domain.BookingActive)

	w := doAuthedJSON(t, router, http.MethodGet, "/bookings", client1.Token, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var resp dto.BookingListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items count = %d, want 2, body: %s", len(resp.Items), w.Body.String())
	}
	seen := map[string]bool{}
	for _, item := range resp.Items {
		seen[item.ID] = true
		if item.ID == "booking-other" {
			t.Fatalf("response leaked another client's booking: %+v", item)
		}
	}
	if !seen["booking-mine-1"] || !seen["booking-mine-2"] {
		t.Fatalf("response missing own bookings, got items: %+v", resp.Items)
	}
}

func TestBookingsHandler_ListNoAuthHeaderIs401(t *testing.T) {
	router := newTestRouterWithBookings(newFakeAuthRepo(), &fakeSlotsRepo{}, newFakeBookingsRepo())

	w := doAuthedJSON(t, router, http.MethodGet, "/bookings", "", nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body: %s", w.Code, w.Body.String())
	}
}

// ---- getBooking ----

func TestBookingsHandler_GetOwnBookingIs200(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "get-owner@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)

	w := doAuthedJSON(t, router, http.MethodGet, "/bookings/booking-1", owner.Token, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var got dto.Booking
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "booking-1" {
		t.Fatalf("booking.id = %q, want booking-1", got.ID)
	}
}

func TestBookingsHandler_GetOtherClientsBookingIs403NotFound(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "get-owner2@example.com")
	other := mustRegister(t, router, "get-other2@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)

	w := doAuthedJSON(t, router, http.MethodGet, "/bookings/booking-1", other.Token, nil)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (NFR-8: not 404), body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeForbidden {
		t.Fatalf("code = %q, want %q", body.Code, CodeForbidden)
	}
}

func TestBookingsHandler_GetNonexistentBookingIs404(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	client := mustRegister(t, router, "get-missing@example.com")

	w := doAuthedJSON(t, router, http.MethodGet, "/bookings/no-such-booking", client.Token, nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeNotFound {
		t.Fatalf("code = %q, want %q", body.Code, CodeNotFound)
	}
}

// ---- cancelBooking ----

func TestBookingsHandler_CancelEarlyReturnsCancelledStatus(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "cancel-early@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)
	bookingsRepo.cancelResult = sampleBooking("booking-1", owner.Client.ID, domain.BookingCancelled)

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/cancel", owner.Token, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var got dto.Booking
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != "cancelled" {
		t.Fatalf("status = %q, want cancelled", got.Status)
	}
}

func TestBookingsHandler_CancelLateReturnsLateCancelStatus(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "cancel-late@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)
	bookingsRepo.cancelResult = sampleBooking("booking-1", owner.Client.ID, domain.BookingLateCancel)

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/cancel", owner.Token, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var got dto.Booking
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != "late_cancel" {
		t.Fatalf("status = %q, want late_cancel", got.Status)
	}
}

func TestBookingsHandler_CancelSlotStartedIs422(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "cancel-started@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)
	bookingsRepo.cancelErr = domain.ErrSlotStarted

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/cancel", owner.Token, nil)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeSlotStarted {
		t.Fatalf("code = %q, want %q", body.Code, CodeSlotStarted)
	}
}

func TestBookingsHandler_CancelAlreadyCancelledIs409(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "cancel-already@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingCancelled)
	bookingsRepo.cancelErr = domain.ErrAlreadyCancelled

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/cancel", owner.Token, nil)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeAlreadyCancelled {
		t.Fatalf("code = %q, want %q", body.Code, CodeAlreadyCancelled)
	}
}

func TestBookingsHandler_CancelOtherClientsBookingIs403(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "cancel-owner@example.com")
	other := mustRegister(t, router, "cancel-other@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/cancel", other.Token, nil)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeForbidden {
		t.Fatalf("code = %q, want %q", body.Code, CodeForbidden)
	}
}

// ---- submitRating ----

func TestBookingsHandler_SubmitRatingSuccess(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "rate-ok@example.com")
	pastBooking := sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)
	bookingsRepo.bookings["booking-1"] = pastBooking
	comment := "Было очень вкусно!"
	rated := pastBooking
	rated.Rating = &domain.Rating{Stars: 5, Comment: &comment, CreatedAt: time.Now()}
	bookingsRepo.ratingResult = rated

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/rating", owner.Token, dto.SubmitRatingRequest{
		Stars:   5,
		Comment: &comment,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var got dto.Booking
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Rating == nil || got.Rating.Stars != 5 {
		t.Fatalf("rating = %+v, want stars=5", got.Rating)
	}
}

// TestBookingsHandler_SubmitRatingStarsOutOfRangeIs400 exercises the contract-documented
// "stars: minimum 1, maximum 5" constraint (bookings/models.yaml SubmitRatingRequest). Nothing
// in the HTTP handler or dto.SubmitRatingRequest validates this range before the repo is called
// — the real bookings_repo.go only checks it deep inside SubmitRating's SQL transaction (see
// report). This test's fake repo replicates that same check so the case is still exercisable
// without a real database; it demonstrates the wiring (ErrInvalidRating -> 400) works, but does
// NOT prove the handler itself guards against it.
func TestBookingsHandler_SubmitRatingStarsOutOfRangeIs400(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "rate-outofrange@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/rating", owner.Token, dto.SubmitRatingRequest{
		Stars: 6,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeBadRequest {
		t.Fatalf("code = %q, want %q", body.Code, CodeBadRequest)
	}
}

func TestBookingsHandler_SubmitRatingNotRatableIs422(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "rate-notratable@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)
	bookingsRepo.ratingErr = domain.ErrNotRatable

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/rating", owner.Token, dto.SubmitRatingRequest{
		Stars: 5,
	})

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeNotRatable {
		t.Fatalf("code = %q, want %q", body.Code, CodeNotRatable)
	}
}

func TestBookingsHandler_SubmitRatingAlreadyRatedIs409(t *testing.T) {
	authRepo := newFakeAuthRepo()
	bookingsRepo := newFakeBookingsRepo()
	router := newTestRouterWithBookings(authRepo, &fakeSlotsRepo{}, bookingsRepo)
	owner := mustRegister(t, router, "rate-already@example.com")
	bookingsRepo.bookings["booking-1"] = sampleBooking("booking-1", owner.Client.ID, domain.BookingActive)
	bookingsRepo.ratingErr = domain.ErrAlreadyRated

	w := doAuthedJSON(t, router, http.MethodPost, "/bookings/booking-1/rating", owner.Token, dto.SubmitRatingRequest{
		Stars: 5,
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	// bookings/api.yaml documents this exact code in its 409 example for submitRating.
	if body.Code != "already_rated" {
		t.Fatalf("code = %q, want %q", body.Code, "already_rated")
	}
}
