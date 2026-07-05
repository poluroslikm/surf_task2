package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chef-stol/backend/internal/domain"
)

// SlotForceCancelRepo is the persistence port used by InternalService (FEAT-05).
type SlotForceCancelRepo interface {
	ForceCancel(ctx context.Context, slotID, reason string) ([]domain.Booking, error)
}

// InternalService backs the dev-only /internal/slots/{slotId}/force-cancel endpoint (FEAT-05,
// FR-17/18/22) — not part of api/, exists purely so studio-initiated cancellation can be
// exercised at all in the absence of an owner/admin panel.
type InternalService struct {
	slots SlotForceCancelRepo
	push  *PushService
}

func NewInternalService(slots SlotForceCancelRepo, push *PushService) *InternalService {
	return &InternalService{slots: slots, push: push}
}

// ForceCancelSlot cascades a studio-initiated cancellation. The DB transaction (slot ->
// cancelled, its active/late_cancel bookings -> cancelled_by_studio) is entirely inside the
// repo call and has already committed by the time it returns; pushes are sent only afterwards,
// so a push failure can never roll back the cancellation — delivery is best-effort.
func (s *InternalService) ForceCancelSlot(ctx context.Context, slotID, reason string) error {
	affected, err := s.slots.ForceCancel(ctx, slotID, reason)
	if err != nil {
		return err
	}
	for _, b := range affected {
		payload, err := json.Marshal(map[string]string{
			"title": "Класс отменён",
			"body": fmt.Sprintf("Класс «%s» %s отменён студией. Причина: %s",
				b.Slot.Program.Name, b.Slot.StartAt.Format("02.01.2006 15:04"), reason),
		})
		if err != nil {
			continue
		}
		s.push.SendToClient(ctx, b.ClientID, payload)
	}
	return nil
}
