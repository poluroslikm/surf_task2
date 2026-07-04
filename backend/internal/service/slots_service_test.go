package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chef-stol/backend/internal/domain"
)

// fakeSlotsRepo is a map-backed in-memory stand-in for SlotsRepo, recording the from/to it was
// last called with so tests can assert on the defaulting logic in SlotsService.List.
type fakeSlotsRepo struct {
	slots         []domain.Slot
	calledFrom    time.Time
	calledTo      time.Time
	listWasCalled bool
}

func (f *fakeSlotsRepo) List(ctx context.Context, from, to time.Time) ([]domain.Slot, error) {
	f.listWasCalled = true
	f.calledFrom = from
	f.calledTo = to
	return f.slots, nil
}

func (f *fakeSlotsRepo) Get(ctx context.Context, id string) (domain.Slot, error) {
	for _, s := range f.slots {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Slot{}, domain.ErrSlotNotFound
}

const timeTolerance = 5 * time.Second

func TestSlotsService_ListDefaultsBothBounds(t *testing.T) {
	repo := &fakeSlotsRepo{}
	svc := NewSlotsService(repo)

	before := time.Now()
	if _, err := svc.List(context.Background(), nil, nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	after := time.Now()

	if !repo.listWasCalled {
		t.Fatal("repo.List was not called")
	}
	if repo.calledFrom.Before(before.Add(-timeTolerance)) || repo.calledFrom.After(after.Add(timeTolerance)) {
		t.Fatalf("from = %v, want close to now (%v .. %v)", repo.calledFrom, before, after)
	}
	wantTo := repo.calledFrom.Add(defaultWindow)
	if diff := repo.calledTo.Sub(wantTo); diff < -timeTolerance || diff > timeTolerance {
		t.Fatalf("to = %v, want from+7d = %v", repo.calledTo, wantTo)
	}
}

func TestSlotsService_ListExplicitDateToPassesThroughFromStillDefaults(t *testing.T) {
	repo := &fakeSlotsRepo{}
	svc := NewSlotsService(repo)

	explicitTo := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Now()
	if _, err := svc.List(context.Background(), nil, &explicitTo); err != nil {
		t.Fatalf("List: %v", err)
	}
	after := time.Now()

	if !repo.calledTo.Equal(explicitTo) {
		t.Fatalf("to = %v, want explicit %v passed straight through", repo.calledTo, explicitTo)
	}
	if repo.calledFrom.Before(before.Add(-timeTolerance)) || repo.calledFrom.After(after.Add(timeTolerance)) {
		t.Fatalf("from = %v, want close to now (%v .. %v)", repo.calledFrom, before, after)
	}
}

func TestSlotsService_ListExplicitDateFrom(t *testing.T) {
	repo := &fakeSlotsRepo{}
	svc := NewSlotsService(repo)

	explicitFrom := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := svc.List(context.Background(), &explicitFrom, nil); err != nil {
		t.Fatalf("List: %v", err)
	}

	if !repo.calledFrom.Equal(explicitFrom) {
		t.Fatalf("from = %v, want explicit %v", repo.calledFrom, explicitFrom)
	}
	wantTo := explicitFrom.Add(defaultWindow)
	if !repo.calledTo.Equal(wantTo) {
		t.Fatalf("to = %v, want from+7d = %v", repo.calledTo, wantTo)
	}
}

func TestSlotsService_GetDelegatesAndPropagatesNotFound(t *testing.T) {
	want := domain.Slot{ID: "slot-1", Price: 1500}
	repo := &fakeSlotsRepo{slots: []domain.Slot{want}}
	svc := NewSlotsService(repo)

	got, err := svc.Get(context.Background(), "slot-1")
	if err != nil {
		t.Fatalf("Get(existing): %v", err)
	}
	if got.ID != want.ID || got.Price != want.Price {
		t.Fatalf("Get(existing) = %+v, want %+v", got, want)
	}

	_, err = svc.Get(context.Background(), "missing")
	if !errors.Is(err, domain.ErrSlotNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrSlotNotFound", err)
	}
}
