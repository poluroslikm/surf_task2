package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
	"github.com/chef-stol/backend/internal/service"
)

// fakePushRepo is a slice-backed stand-in for service.PushRepo, used to drive the real
// PushHandler/PushService/InternalHandler/InternalService under httptest without any external
// dependency (mirrors fakeAuthRepo/fakeSlotsRepo in fakes_test.go).
type fakePushRepo struct {
	subs []domain.PushSubscription
}

func (f *fakePushRepo) Upsert(ctx context.Context, clientID, endpoint, p256dh, auth string) (domain.PushSubscription, error) {
	for i, s := range f.subs {
		if s.Endpoint == endpoint {
			f.subs[i].ClientID = clientID
			f.subs[i].P256dh = p256dh
			f.subs[i].Auth = auth
			return f.subs[i], nil
		}
	}
	sub := domain.PushSubscription{
		ID:        fmt.Sprintf("push-%d", len(f.subs)+1),
		ClientID:  clientID,
		Endpoint:  endpoint,
		P256dh:    p256dh,
		Auth:      auth,
		CreatedAt: time.Now(),
	}
	f.subs = append(f.subs, sub)
	return sub, nil
}

func (f *fakePushRepo) ListByClient(ctx context.Context, clientID string) ([]domain.PushSubscription, error) {
	var out []domain.PushSubscription
	for _, s := range f.subs {
		if s.ClientID == clientID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakePushRepo) Delete(ctx context.Context, endpoint string) error {
	for i, s := range f.subs {
		if s.Endpoint == endpoint {
			f.subs = append(f.subs[:i], f.subs[i+1:]...)
			return nil
		}
	}
	return nil
}

// fakeForceCancelRepo is a stand-in for service.SlotForceCancelRepo, recording the slotID/reason
// it was called with and returning whatever err/affected bookings the test configured.
type fakeForceCancelRepo struct {
	err                        error
	affected                   []domain.Booking
	calledSlotID, calledReason string
	called                     bool
}

func (f *fakeForceCancelRepo) ForceCancel(ctx context.Context, slotID, reason string) ([]domain.Booking, error) {
	f.called = true
	f.calledSlotID = slotID
	f.calledReason = reason
	if f.err != nil {
		return nil, f.err
	}
	return f.affected, nil
}

// newPushInternalTestRouter wires the real Push and Internal handlers on top of fake repos,
// alongside the real Auth handler (needed to obtain a bearer token for RequireAuth-guarded
// /push/subscriptions). Deliberately a separate constructor from fakes_test.go's newTestRouter
// rather than modifying it, since that one is also used by slots/auth handler tests that don't
// need push/internal wiring.
func newPushInternalTestRouter(authRepo *fakeAuthRepo, pushRepo service.PushRepo, forceCancelRepo service.SlotForceCancelRepo, internalToken string) http.Handler {
	authSvc := service.NewAuthService(authRepo, time.Hour)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pushSvc := service.NewPushService(pushRepo, service.NewPushSender("pub", "priv", "mailto:test@example.com"), logger)
	internalSvc := service.NewInternalService(forceCancelRepo, pushSvc)
	h := Handlers{
		Auth:     NewAuthHandler(authSvc),
		Push:     NewPushHandler(pushSvc),
		Internal: NewInternalHandler(internalSvc),
	}
	return NewRouter(h, authSvc, internalToken, logger)
}

// registerAndAuthenticate registers a fresh client through the real /auth/register handler and
// returns its bearer token, so tests exercise the full auth -> push flow rather than
// hand-crafting a session.
func registerAndAuthenticate(t *testing.T, router http.Handler, email string) string {
	t.Helper()
	resp := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    email,
		Password: "password1",
	})
	if resp.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201, body: %s", resp.Code, resp.Body.String())
	}
	var auth dto.AuthResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &auth); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	return auth.Token
}

func TestPushHandler_RegisterSuccessEchoesSubscription(t *testing.T) {
	authRepo := newFakeAuthRepo()
	pushRepo := &fakePushRepo{}
	router := newPushInternalTestRouter(authRepo, pushRepo, &fakeForceCancelRepo{}, "test-internal-token")
	token := registerAndAuthenticate(t, router, "pushuser@example.com")

	body := dto.PushSubscriptionRequest{
		Endpoint: "https://push.example.com/sub1",
		Keys: dto.PushSubscriptionKeys{
			P256dh: "some-p256dh-key",
			Auth:   "some-auth-secret",
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/push/subscriptions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var got dto.PushSubscriptionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Endpoint != body.Endpoint || got.Keys.P256dh != body.Keys.P256dh || got.Keys.Auth != body.Keys.Auth {
		t.Fatalf("response = %+v, want echoed subscription %+v", got, body)
	}
	if len(pushRepo.subs) != 1 {
		t.Fatalf("pushRepo.subs = %d, want 1 stored subscription", len(pushRepo.subs))
	}
}

func TestPushHandler_RegisterMalformedEndpointIs400(t *testing.T) {
	authRepo := newFakeAuthRepo()
	pushRepo := &fakePushRepo{}
	router := newPushInternalTestRouter(authRepo, pushRepo, &fakeForceCancelRepo{}, "test-internal-token")
	token := registerAndAuthenticate(t, router, "pushuser2@example.com")

	body := dto.PushSubscriptionRequest{
		Endpoint: "/relative-not-absolute",
		Keys: dto.PushSubscriptionKeys{
			P256dh: "some-p256dh-key",
			Auth:   "some-auth-secret",
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/push/subscriptions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body: %s", w.Code, w.Body.String())
	}
	if len(pushRepo.subs) != 0 {
		t.Fatalf("pushRepo.subs = %d, want 0 (rejected before reaching the repo)", len(pushRepo.subs))
	}
}

func TestPushHandler_RegisterNoAuthHeaderIs401(t *testing.T) {
	authRepo := newFakeAuthRepo()
	pushRepo := &fakePushRepo{}
	router := newPushInternalTestRouter(authRepo, pushRepo, &fakeForceCancelRepo{}, "test-internal-token")

	body := dto.PushSubscriptionRequest{
		Endpoint: "https://push.example.com/sub1",
		Keys: dto.PushSubscriptionKeys{
			P256dh: "some-p256dh-key",
			Auth:   "some-auth-secret",
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/push/subscriptions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body: %s", w.Code, w.Body.String())
	}
}
