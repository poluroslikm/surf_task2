package service

import (
	"context"
	"fmt"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/chef-stol/backend/internal/domain"
)

// pushDeleter is the subset of PushRepo needed to clean up a dead subscription after a failed
// send. Kept minimal so PushSender doesn't depend on the full repo surface.
type pushDeleter interface {
	Delete(ctx context.Context, endpoint string) error
}

// PushSender wraps webpush-go with this project's VAPID key pair (FEAT-03). The dev key pair is
// fixed for this delivery (see backend/.env.example) so the frontend engineer's hardcoded
// public key matches what this backend signs with.
type PushSender struct {
	vapidPublicKey  string
	vapidPrivateKey string
	vapidSubject    string
}

func NewPushSender(vapidPublicKey, vapidPrivateKey, vapidSubject string) *PushSender {
	return &PushSender{
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
		vapidSubject:    vapidSubject,
	}
}

// Send delivers payload to sub via VAPID-signed Web Push. A 404/410 response means the browser
// has invalidated the subscription — this is cleanup (delete it via repo), not an error for the
// caller to retry or propagate; any other non-2xx response is returned as an error.
func (s *PushSender) Send(ctx context.Context, repo pushDeleter, sub domain.PushSubscription, payload []byte) error {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}, &webpush.Options{
		Subscriber:      s.vapidSubject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             60,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return repo.Delete(ctx, sub.Endpoint)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("push send failed: %s", resp.Status)
	}
	return nil
}
