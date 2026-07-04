package service

import (
	"context"

	"github.com/chef-stol/backend/internal/domain"
)

// BookingsRepo is the persistence port used by BookingsService. Ownership checks for
// cancel/rating are enforced inside the repo (same transaction as the row lock) rather than
// here, to avoid a TOCTOU window between an ownership read and the mutating statement.
type BookingsRepo interface {
	Create(ctx context.Context, clientID, slotID string, equipment domain.EquipmentChoice) (domain.Booking, error)
	List(ctx context.Context, clientID string) ([]domain.Booking, error)
	Get(ctx context.Context, id string) (domain.Booking, error)
	Cancel(ctx context.Context, bookingID, clientID string) (domain.Booking, error)
	SubmitRating(ctx context.Context, bookingID, clientID string, stars int, comment *string) (domain.Booking, error)
}

type BookingsService struct {
	repo BookingsRepo
}

func NewBookingsService(repo BookingsRepo) *BookingsService {
	return &BookingsService{repo: repo}
}

func (s *BookingsService) Create(ctx context.Context, clientID, slotID string, equipment domain.EquipmentChoice) (domain.Booking, error) {
	return s.repo.Create(ctx, clientID, slotID, equipment)
}

func (s *BookingsService) List(ctx context.Context, clientID string) ([]domain.Booking, error) {
	return s.repo.List(ctx, clientID)
}

// Get enforces per-client ownership (NFR-8): getBooking's contract documents 403 for another
// client's booking, not 404 — so existence and ownership are deliberately not conflated here.
func (s *BookingsService) Get(ctx context.Context, clientID, id string) (domain.Booking, error) {
	b, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Booking{}, err
	}
	if b.ClientID != clientID {
		return domain.Booking{}, domain.ErrForbidden
	}
	return b, nil
}

func (s *BookingsService) Cancel(ctx context.Context, clientID, id string) (domain.Booking, error) {
	return s.repo.Cancel(ctx, id, clientID)
}

func (s *BookingsService) SubmitRating(ctx context.Context, clientID, id string, stars int, comment *string) (domain.Booking, error) {
	return s.repo.SubmitRating(ctx, id, clientID, stars, comment)
}
