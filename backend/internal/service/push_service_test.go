package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/chef-stol/backend/internal/domain"
)

// fakePushRepo is a slice-backed in-memory stand-in for PushRepo, recording every Upsert call
// (clientID/endpoint/p256dh/auth) so tests can assert the service actually forwarded the right
// arguments rather than silently skipping the call, and every Delete call for the 404/410
// cleanup path exercised in TestPushService_SendToClient.
type fakePushRepo struct {
	subs        []domain.PushSubscription
	upsertCalls []upsertCall
	deleteCalls []string
	nextID      int
}

type upsertCall struct {
	clientID, endpoint, p256dh, auth string
}

func (f *fakePushRepo) Upsert(ctx context.Context, clientID, endpoint, p256dh, auth string) (domain.PushSubscription, error) {
	f.upsertCalls = append(f.upsertCalls, upsertCall{clientID, endpoint, p256dh, auth})
	for i, s := range f.subs {
		if s.Endpoint == endpoint {
			f.subs[i].ClientID = clientID
			f.subs[i].P256dh = p256dh
			f.subs[i].Auth = auth
			return f.subs[i], nil
		}
	}
	f.nextID++
	sub := domain.PushSubscription{
		ID:        fmt.Sprintf("push-%d", f.nextID),
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
	f.deleteCalls = append(f.deleteCalls, endpoint)
	for i, s := range f.subs {
		if s.Endpoint == endpoint {
			f.subs = append(f.subs[:i], f.subs[i+1:]...)
			break
		}
	}
	return nil
}

// validRecipientKeys returns a syntactically valid P-256 public key (p256dh) and a 16-byte
// auth secret, base64url-encoded, the same shape a real browser subscription would supply.
// GenerateVAPIDKeys happens to produce exactly the "base64url(marshaled P-256 point)" format
// needed for p256dh, so it's reused here purely as a convenient valid-EC-point generator.
func validRecipientKeys(t *testing.T) (p256dh, auth string) {
	t.Helper()
	_, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("GenerateVAPIDKeys: %v", err)
	}
	authBytes := make([]byte, 16)
	if _, err := rand.Read(authBytes); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return pub, base64.RawURLEncoding.EncodeToString(authBytes)
}

func TestPushService_RegisterSubscription_EmptyFieldsRejected(t *testing.T) {
	p256dh, auth := validRecipientKeys(t)
	cases := []struct {
		name           string
		endpoint, p, a string
	}{
		{"empty endpoint", "", p256dh, auth},
		{"empty p256dh", "https://push.example.com/sub1", "", auth},
		{"empty auth", "https://push.example.com/sub1", p256dh, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakePushRepo{}
			svc := NewPushService(repo, NewPushSender("pub", "priv", "mailto:test@example.com"), nil)
			_, err := svc.RegisterSubscription(context.Background(), "client-1", tc.endpoint, tc.p, tc.a)
			if err != domain.ErrInvalidPushSubscription {
				t.Fatalf("err = %v, want ErrInvalidPushSubscription", err)
			}
			if len(repo.upsertCalls) != 0 {
				t.Fatalf("repo.Upsert should not have been called, got %d calls", len(repo.upsertCalls))
			}
		})
	}
}

func TestPushService_RegisterSubscription_RelativeURLRejected(t *testing.T) {
	p256dh, auth := validRecipientKeys(t)
	repo := &fakePushRepo{}
	svc := NewPushService(repo, NewPushSender("pub", "priv", "mailto:test@example.com"), nil)

	_, err := svc.RegisterSubscription(context.Background(), "client-1", "/foo", p256dh, auth)
	if err != domain.ErrInvalidPushSubscription {
		t.Fatalf("err = %v, want ErrInvalidPushSubscription", err)
	}
	if len(repo.upsertCalls) != 0 {
		t.Fatalf("repo.Upsert should not have been called for a relative URL, got %d calls", len(repo.upsertCalls))
	}
}

func TestPushService_RegisterSubscription_WellFormedSucceedsAndCallsUpsert(t *testing.T) {
	p256dh, auth := validRecipientKeys(t)
	repo := &fakePushRepo{}
	svc := NewPushService(repo, NewPushSender("pub", "priv", "mailto:test@example.com"), nil)

	sub, err := svc.RegisterSubscription(context.Background(), "client-1", "https://push.example.com/sub1", p256dh, auth)
	if err != nil {
		t.Fatalf("RegisterSubscription: %v", err)
	}
	if sub.Endpoint != "https://push.example.com/sub1" || sub.P256dh != p256dh || sub.Auth != auth {
		t.Fatalf("returned subscription = %+v, want endpoint/p256dh/auth echoed back", sub)
	}
	if len(repo.upsertCalls) != 1 {
		t.Fatalf("repo.Upsert calls = %d, want 1", len(repo.upsertCalls))
	}
	got := repo.upsertCalls[0]
	if got.clientID != "client-1" || got.endpoint != "https://push.example.com/sub1" || got.p256dh != p256dh || got.auth != auth {
		t.Fatalf("repo.Upsert called with %+v, want clientID=client-1 endpoint=https://push.example.com/sub1", got)
	}
}

func TestPushService_RegisterSubscription_ReRegisterSameEndpointUpserts(t *testing.T) {
	p256dhA, authA := validRecipientKeys(t)
	p256dhB, authB := validRecipientKeys(t)
	repo := &fakePushRepo{}
	svc := NewPushService(repo, NewPushSender("pub", "priv", "mailto:test@example.com"), nil)

	if _, err := svc.RegisterSubscription(context.Background(), "client-1", "https://push.example.com/sub1", p256dhA, authA); err != nil {
		t.Fatalf("first RegisterSubscription: %v", err)
	}
	// Re-registering the same endpoint (e.g. new keys, or a different owning client on a
	// shared device) must still succeed — it's an upsert, never a conflict.
	sub, err := svc.RegisterSubscription(context.Background(), "client-2", "https://push.example.com/sub1", p256dhB, authB)
	if err != nil {
		t.Fatalf("re-registration should succeed (upsert), got error: %v", err)
	}
	if sub.ClientID != "client-2" || sub.P256dh != p256dhB || sub.Auth != authB {
		t.Fatalf("re-registered subscription = %+v, want it updated to client-2/new keys", sub)
	}
	if len(repo.upsertCalls) != 2 {
		t.Fatalf("repo.Upsert calls = %d, want 2 (one per registration)", len(repo.upsertCalls))
	}
	if len(repo.subs) != 1 {
		t.Fatalf("repo should still hold exactly 1 subscription for this endpoint (upsert, not insert), got %d", len(repo.subs))
	}
}

// captureLogger returns a logger writing to buf so tests can assert on which
// endpoints/clients were logged as send failures, without depending on internal call counts
// that PushService/PushSender don't expose directly.
func captureLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, nil))
}

// TestPushService_SendToClient_OneFailureDoesNotStopTheOthers is the important fan-out test:
// PushSender has no interface/seam to inject a fake transport, so this drives the real
// *PushSender end-to-end. The first subscription has syntactically invalid key material, which
// fails locally during Web Push payload encryption (decodeSubscriptionKey/ECDH, see
// webpush-go's SendNotificationWithContext) before any network I/O is attempted — so this stays
// fully local/deterministic. The second subscription has valid keys and points at a local
// httptest.Server, proving that the first subscription's failure did not stop delivery to the
// second (best-effort fan-out).
func TestPushService_SendToClient_OneFailureDoesNotStopTheOthers(t *testing.T) {
	var receivedRequests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequests++
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	validP256dh, validAuth := validRecipientKeys(t)

	repo := &fakePushRepo{subs: []domain.PushSubscription{
		{ID: "sub-bad", ClientID: "client-1", Endpoint: "https://push.example.com/bad", P256dh: "not-valid-base64!!!", Auth: "also-not-valid!!!"},
		{ID: "sub-good", ClientID: "client-1", Endpoint: srv.URL + "/good", P256dh: validP256dh, Auth: validAuth},
	}}

	var logBuf bytes.Buffer
	vapidPriv, vapidPub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("GenerateVAPIDKeys: %v", err)
	}
	svc := NewPushService(repo, NewPushSender(vapidPub, vapidPriv, "mailto:test@example.com"), captureLogger(&logBuf))

	svc.SendToClient(context.Background(), "client-1", []byte(`{"title":"hi"}`))

	if receivedRequests != 1 {
		t.Fatalf("httptest server received %d requests, want exactly 1 (only the valid subscription should reach the network)", receivedRequests)
	}
	logged := logBuf.String()
	if !strings.Contains(logged, "https://push.example.com/bad") {
		t.Fatalf("expected the failing subscription's endpoint to be logged, got log output: %s", logged)
	}
	if strings.Contains(logged, srv.URL+"/good") {
		t.Fatalf("the successful subscription should not have logged an error, got log output: %s", logged)
	}
	if len(repo.deleteCalls) != 0 {
		t.Fatalf("no subscription returned 404/410, repo.Delete should not have been called, got %v", repo.deleteCalls)
	}
}

// TestPushService_SendToClient_410ResponseDeletesSubscription confirms the dead-subscription
// cleanup path: a 410 Gone response from the push endpoint results in the subscription being
// deleted from the repo, matching PushSender.Send's documented contract.
func TestPushService_SendToClient_410ResponseDeletesSubscription(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
	}))
	defer srv.Close()

	validP256dh, validAuth := validRecipientKeys(t)
	repo := &fakePushRepo{subs: []domain.PushSubscription{
		{ID: "sub-dead", ClientID: "client-1", Endpoint: srv.URL + "/dead", P256dh: validP256dh, Auth: validAuth},
	}}

	vapidPriv, vapidPub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("GenerateVAPIDKeys: %v", err)
	}
	svc := NewPushService(repo, NewPushSender(vapidPub, vapidPriv, "mailto:test@example.com"), slog.New(slog.NewTextHandler(new(bytes.Buffer), nil)))

	svc.SendToClient(context.Background(), "client-1", []byte(`{"title":"hi"}`))

	if len(repo.deleteCalls) != 1 || repo.deleteCalls[0] != srv.URL+"/dead" {
		t.Fatalf("repo.deleteCalls = %v, want exactly one delete for %s", repo.deleteCalls, srv.URL+"/dead")
	}
}
