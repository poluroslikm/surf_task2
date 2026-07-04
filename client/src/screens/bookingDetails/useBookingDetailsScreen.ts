import { useCallback, useEffect, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { bookingsApi } from '../../api/bookingsApi'
import { ScreenState } from '../../core/screenState'
import { ApiRequestError, NetworkError, GENERIC_LOAD_NETWORK_ERROR, GENERIC_SERVER_ERROR } from '../../core/errors'
import type { Booking } from '../../api/types'
import type { CancelOutcome } from '../../components/CancelConfirmSheet'

function loadErrorMessage(err: unknown): string {
  if (err instanceof NetworkError) return GENERIC_LOAD_NETWORK_ERROR
  if (err instanceof ApiRequestError && err.status >= 500) return GENERIC_SERVER_ERROR
  return GENERIC_LOAD_NETWORK_ERROR
}

// SCR-006 + LOGIC-002 + LOGIC-005.
export function useBookingDetailsScreen() {
  const { bookingId } = useParams<{ bookingId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const [screen, setScreen] = useState<ScreenState<Booking>>(ScreenState.loading())
  const [cancelSheetOpen, setCancelSheetOpen] = useState(false)
  const [snack, setSnack] = useState<string | null>(null)

  // `showLoadingSkeleton=false` is used for the post-cancel refetch: the state diagram in
  // SCR-006 goes straight from ContentActive to ContentCancelled (no intermediate Loading), and
  // a background-refetch failure must not blow away the content/snack already on screen
  // (LOGIC-005 AC-007's "Refreshing never falls back to Error" rule, applied here to a
  // single-record screen instead of a list).
  const load = useCallback(
    async (showLoadingSkeleton: boolean) => {
      if (!bookingId) {
        // No navigation param at all — nothing to load, not a "some record was empty" case
        // (NFR-8 / "only own data" — an absent id is treated the same as a 404).
        setScreen(ScreenState.error(GENERIC_LOAD_NETWORK_ERROR))
        return
      }
      if (showLoadingSkeleton) setScreen(ScreenState.loading())
      try {
        const booking = await bookingsApi.getBooking(bookingId)
        setScreen(ScreenState.content(booking))
      } catch (err) {
        if (err instanceof ApiRequestError && err.status === 401) return
        if (showLoadingSkeleton) setScreen(ScreenState.error(loadErrorMessage(err)))
      }
    },
    [bookingId],
  )

  useEffect(() => {
    void load(true)
  }, [load])

  // AC-005: SCR-007 navigates back here after a successful submitRating and this screen (as the
  // parent that outlives the child screen) is the one that shows the "Спасибо за оценку!" snack —
  // signalled via navigation state so it fires exactly once, not on every refetch.
  useEffect(() => {
    const state = location.state as { justRated?: boolean } | null
    if (state?.justRated) {
      showSnack('Спасибо за оценку!')
      navigate(location.pathname, { replace: true, state: {} })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.state])

  function showSnack(text: string) {
    setSnack(text)
    setTimeout(() => setSnack(null), 4000)
  }

  // BS-003 -> SCR-006 handoff (foundations §6.2): the sheet only reports the outcome, the parent
  // screen shows the resulting snack and refetches so status/CTA reflect the server's answer.
  function handleCancelOutcome(outcome: CancelOutcome) {
    setCancelSheetOpen(false)
    if (outcome.kind === 'cancelled') {
      showSnack(
        outcome.status === 'cancelled'
          ? 'Запись отменена'
          : 'Поздняя отмена: продукты на класс уже закуплены. Штраф не взимается.',
      )
    } else {
      showSnack(outcome.code === 'slot_started' ? 'Класс уже начался — отменить запись нельзя.' : 'Запись уже отменена.')
    }
    void load(false)
  }

  return {
    bookingId,
    screen,
    snack,
    cancelSheetOpen,
    openCancelSheet: () => setCancelSheetOpen(true),
    closeCancelSheet: () => setCancelSheetOpen(false),
    handleCancelOutcome,
    retryAfterError: () => void load(true),
  }
}
