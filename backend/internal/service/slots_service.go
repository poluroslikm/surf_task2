package service

import (
	"context"
	"time"

	"github.com/chef-stol/backend/internal/domain"
)

// defaultWindow mirrors FR-3/R-027: the client sees a 7-day horizon by default.
const defaultWindow = 7 * 24 * time.Hour

type SlotsRepo interface {
	List(ctx context.Context, from, to time.Time) ([]domain.Slot, error)
	Get(ctx context.Context, id string) (domain.Slot, error)
}

type SlotsService struct {
	repo SlotsRepo
}

func NewSlotsService(repo SlotsRepo) *SlotsService {
	return &SlotsService{repo: repo}
}

// List applies the FR-3/FR-4/R-027 defaulting rule: dateFrom defaults to now, dateTo
// defaults to dateFrom + 7 days. Filled and studio-cancelled slots are not filtered out
// here (FR-6a, FR-18) — the client decides how to render them.
func (s *SlotsService) List(ctx context.Context, dateFrom, dateTo *time.Time) ([]domain.Slot, error) {
	from := time.Now().UTC()
	if dateFrom != nil {
		from = *dateFrom
	}
	to := from.Add(defaultWindow)
	if dateTo != nil {
		to = *dateTo
	}
	return s.repo.List(ctx, from, to)
}

func (s *SlotsService) Get(ctx context.Context, id string) (domain.Slot, error) {
	return s.repo.Get(ctx, id)
}
