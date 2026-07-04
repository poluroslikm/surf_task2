import { useState } from 'react'
import { bookingsApi } from '../api/bookingsApi'
import { ApiRequestError, NetworkError, GENERIC_NETWORK_ERROR, GENERIC_SERVER_ERROR } from '../core/errors'

// BS-003 §6 / foundations §6: verbatim cancellation-rule text, shown only in the "late cancel"
// preview variant.
const CANCELLATION_RULE_TEXT =
  'Отмена не позднее чем за 24 часа до начала — без предупреждений, место освобождается. Позже — запись всё равно отменится, но продукты на класс уже закуплены. Штраф не взимается.'

export type CancelOutcome =
  | { kind: 'cancelled'; status: 'cancelled' | 'late_cancel' }
  | { kind: 'race'; code: 'slot_started' | 'already_cancelled' }

// BS-003. Critical-confirmation sheet — no backdrop-click / swipe-to-close (foundations §4.3
// exception), only the explicit "Не отменять" button or the ✕. The early/late preview variant is
// a plain comparison of `now` against the already server-computed `free_cancellation_until`
// (LOGIC-002) — never a client recomputation of `slot.start_at - 24h`. Whatever the preview
// showed, the final outcome (and the snack/badge shown on SCR-006 afterwards) is driven solely by
// the `status` field `cancelBooking` actually returns.
export function CancelConfirmSheet({
  bookingId,
  freeCancellationUntil,
  onClose,
  onOutcome,
}: {
  bookingId: string
  freeCancellationUntil: string
  onClose: () => void
  onOutcome: (outcome: CancelOutcome) => void
}) {
  const [isLate] = useState(() => new Date().getTime() >= new Date(freeCancellationUntil).getTime())
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function confirmCancel() {
    setSubmitting(true)
    setError(null)
    try {
      const booking = await bookingsApi.cancelBooking(bookingId)
      if (booking.status === 'cancelled' || booking.status === 'late_cancel') {
        onOutcome({ kind: 'cancelled', status: booking.status })
      }
      // Any other status here would mean the server didn't actually cancel — shouldn't happen
      // per the contract, and there's nothing sensible to show if it did, so we simply stop.
    } catch (err) {
      if (err instanceof ApiRequestError && err.status === 401) return // httpClient already cleared the session; App redirects.

      if (err instanceof ApiRequestError && err.status === 422 && err.body.code === 'slot_started') {
        onOutcome({ kind: 'race', code: 'slot_started' })
        return
      }
      if (err instanceof ApiRequestError && err.status === 409 && err.body.code === 'already_cancelled') {
        onOutcome({ kind: 'race', code: 'already_cancelled' })
        return
      }

      // BS-003 AC-N01: network/5xx keeps the sheet open with its own snack; the confirm button
      // unblocks so the client can retry with the same tap.
      setError(err instanceof NetworkError ? GENERIC_NETWORK_ERROR : GENERIC_SERVER_ERROR)
      setSubmitting(false)
    }
  }

  return (
    <div className="sheet-backdrop">
      <div className="sheet">
        <div className="sheet__header">
          <h2>Точно отменить запись?</h2>
          <button className="sheet__close" aria-label="Закрыть" onClick={onClose} disabled={submitting}>
            ✕
          </button>
        </div>

        {isLate && <p className="cancel-sheet__warning">{CANCELLATION_RULE_TEXT}</p>}

        {error && <div className="form-error">{error}</div>}

        <div className="sheet__actions">
          <button className="sheet__action sheet__action--primary" disabled={submitting} onClick={onClose}>
            Не отменять
          </button>
          <button className="sheet__action sheet__action--danger" disabled={submitting} onClick={() => void confirmCancel()}>
            {submitting ? '…' : 'Отменить запись'}
          </button>
        </div>
      </div>
    </div>
  )
}
