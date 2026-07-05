package service

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/chef-stol/backend/internal/domain"
)

// PushRepo is the persistence port used by PushService (FEAT-03).
type PushRepo interface {
	Upsert(ctx context.Context, clientID, endpoint, p256dh, auth string) (domain.PushSubscription, error)
	ListByClient(ctx context.Context, clientID string) ([]domain.PushSubscription, error)
	Delete(ctx context.Context, endpoint string) error
}

type PushService struct {
	repo   PushRepo
	sender *PushSender
	logger *slog.Logger
}

func NewPushService(repo PushRepo, sender *PushSender, logger *slog.Logger) *PushService {
	return &PushService{repo: repo, sender: sender, logger: logger}
}

// RegisterSubscription validates and upserts a Web Push subscription (FEAT-03,
// registerPushSubscription). endpoint must be a non-empty absolute URL
// (push/models.yaml#/components/schemas/PushSubscription); p256dh/auth must be non-empty
// base64url key material — their exact format isn't validated further here.
func (s *PushService) RegisterSubscription(ctx context.Context, clientID, endpoint, p256dh, auth string) (domain.PushSubscription, error) {
	if endpoint == "" || p256dh == "" || auth == "" {
		return domain.PushSubscription{}, domain.ErrInvalidPushSubscription
	}
	u, err := url.ParseRequestURI(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return domain.PushSubscription{}, domain.ErrInvalidPushSubscription
	}
	return s.repo.Upsert(ctx, clientID, endpoint, p256dh, auth)
}

// SendPush delivers a single push message via VAPID-signed Web Push, reused by the reminder
// worker (FEAT-04) and force-cancel (FEAT-05).
func (s *PushService) SendPush(ctx context.Context, sub domain.PushSubscription, payload []byte) error {
	return s.sender.Send(ctx, s.repo, sub, payload)
}

// SendToClient best-effort delivers payload to every subscription belonging to clientID
// (FEAT-04 reminders, FEAT-05 force-cancel). A send failure on one subscription is logged and
// does not stop delivery to the client's other subscriptions, nor is it propagated to the
// caller — this is fire-and-forget by design (see FEAT-04/FEAT-05 task notes).
func (s *PushService) SendToClient(ctx context.Context, clientID string, payload []byte) {
	subs, err := s.repo.ListByClient(ctx, clientID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("push: list subscriptions", "client_id", clientID, "error", err)
		}
		return
	}
	for _, sub := range subs {
		if err := s.SendPush(ctx, sub, payload); err != nil && s.logger != nil {
			s.logger.Error("push: send", "client_id", clientID, "endpoint", sub.Endpoint, "error", err)
		}
	}
}
