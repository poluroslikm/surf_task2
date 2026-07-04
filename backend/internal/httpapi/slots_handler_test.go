package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
)

func TestSlotsHandler_ListNoAuthHeaderIs401(t *testing.T) {
	router := newTestRouter(newFakeAuthRepo(), &fakeSlotsRepo{})

	req := httptest.NewRequest(http.MethodGet, "/slots", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body: %s", w.Code, w.Body.String())
	}
}

func TestSlotsHandler_ListWithValidTokenReturnsSlots(t *testing.T) {
	authRepo := newFakeAuthRepo()
	startAt := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	slotsRepo := &fakeSlotsRepo{
		slots: []domain.Slot{
			{
				ID: "slot-1",
				Program: domain.Program{
					ID:          "prog-1",
					Name:        "Паста ручной работы",
					Difficulty:  domain.DifficultyNovice,
					PhotoURL:    "https://example.com/photo.jpg",
					Ingredients: []string{"мука", "яйца"},
					Allergens:   []string{"глютен"},
				},
				Chef:       domain.Chef{ID: "chef-1", Name: "Иван Иванов"},
				StartAt:    startAt,
				TotalSeats: 12,
				FreeSeats:  10,
				Price:      2500,
				Status:     domain.SlotScheduled,
			},
		},
	}
	router := newTestRouter(authRepo, slotsRepo)

	// Register through the real handler so the token is issued by the real service, exercising
	// the full auth -> slots flow rather than hand-crafting a session.
	regResp := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    "user@example.com",
		Password: "password1",
	})
	if regResp.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201", regResp.Code)
	}
	var auth dto.AuthResponse
	if err := json.Unmarshal(regResp.Body.Bytes(), &auth); err != nil {
		t.Fatalf("decode register response: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/slots", nil)
	req.Header.Set("Authorization", "Bearer "+auth.Token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var listResp dto.SlotListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(listResp.Items))
	}
	got := listResp.Items[0]
	if got.ID != "slot-1" || got.Price != 2500 || got.FreeSeats != 10 || got.Chef.Name != "Иван Иванов" {
		t.Fatalf("slot round-trip mismatch: %+v", got)
	}
	if !got.StartAt.Equal(startAt) {
		t.Fatalf("start_at = %v, want %v", got.StartAt, startAt)
	}
}
