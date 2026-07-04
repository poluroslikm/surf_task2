import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { slotsApi } from '../../api/slotsApi'
import { bookingsApi } from '../../api/bookingsApi'
import { ScreenState } from '../../core/screenState'
import {
  ApiRequestError,
  NetworkError,
  GENERIC_LOAD_NETWORK_ERROR,
  GENERIC_NETWORK_ERROR,
  GENERIC_SERVER_ERROR,
} from '../../core/errors'
import type { Booking, EquipmentChoice, Slot } from '../../api/types'

function loadErrorMessage(err: unknown): string {
  if (err instanceof NetworkError) return GENERIC_LOAD_NETWORK_ERROR
  if (err instanceof ApiRequestError && err.status >= 500) return GENERIC_SERVER_ERROR
  return GENERIC_LOAD_NETWORK_ERROR
}

// api-sequence.md scenario 1 / SCR-004-booking.md "Используемые запросы": 400 bad_request uses
// the server's own `message`; network/5xx use the generic sквозные texts (00-foundations.md §6).
function submitErrorMessage(err: unknown): string {
  if (err instanceof NetworkError) return GENERIC_NETWORK_ERROR
  if (err instanceof ApiRequestError) {
    if (err.status >= 500) return GENERIC_SERVER_ERROR
    return err.body.message || GENERIC_NETWORK_ERROR
  }
  return GENERIC_NETWORK_ERROR
}

const NAV_AWAY_DELAY_MS = 1600

// SCR-004 + LOGIC-005 (action branch). Deliberate deviation from the spec's own "Входные
// данные" decision: this screen re-fetches getSlot(slotId) itself instead of receiving `slot` as
// a nav param from SCR-003 — per this iteration's FE-07/FE-08 rule (re-fetch on both screens, never
// trust anything cached from a previous screen/list).
export function useBookingScreen(slotId: string) {
  const navigate = useNavigate()
  const [slotScreen, setSlotScreen] = useState<ScreenState<Slot>>(ScreenState.loading())
  const [equipmentChoice, setEquipmentChoice] = useState<EquipmentChoice>('own')
  const [submitting, setSubmitting] = useState(false)
  const [snack, setSnack] = useState<string | null>(null)
  const [booking, setBooking] = useState<Booking | null>(null)
  const snackTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const loadSlot = useCallback(() => {
    setSlotScreen(ScreenState.loading())
    slotsApi
      .getSlot(slotId)
      .then((slot) => setSlotScreen(ScreenState.content(slot)))
      .catch((err) => {
        if (err instanceof ApiRequestError && err.status === 401) return
        setSlotScreen(ScreenState.error(loadErrorMessage(err)))
      })
  }, [slotId])

  useEffect(() => {
    loadSlot()
  }, [loadSlot])

  useEffect(
    () => () => {
      if (snackTimer.current) clearTimeout(snackTimer.current)
    },
    [],
  )

  // Copies useSlotsScreen.ts's exact transient-snack mechanism: set, then auto-clear.
  function showSnack(message: string, autoDismissMs = 4000) {
    setSnack(message)
    if (snackTimer.current) clearTimeout(snackTimer.current)
    snackTimer.current = setTimeout(() => setSnack(null), autoDismissMs)
  }

  async function submit() {
    // NFR-6: a second tap while the first request is in flight must not fire another request.
    if (submitting || slotScreen.kind !== 'content') return
    setSubmitting(true)

    try {
      const result = await bookingsApi.createBooking({ slot_id: slotId, equipment_choice: equipmentChoice })
      setBooking(result)
    } catch (err) {
      if (err instanceof ApiRequestError && err.status === 401) return

      // api-sequence.md scenario 1: no waitlist, no auto-retry — the client just returns to the
      // slot list with the fixed snack text, it doesn't stay on the form with an inline error.
      if (err instanceof ApiRequestError && err.status === 409 && err.body.code === 'slot_full') {
        showSnack('Мест больше нет. Список обновлён.', NAV_AWAY_DELAY_MS)
        setTimeout(() => navigate('/', { replace: true }), NAV_AWAY_DELAY_MS)
        return
      }

      if (err instanceof ApiRequestError && err.status === 410 && err.body.code === 'slot_cancelled') {
        showSnack('Класс отменён студией', NAV_AWAY_DELAY_MS)
        setTimeout(() => navigate('/', { replace: true }), NAV_AWAY_DELAY_MS)
        return
      }

      // 400 bad_request / network / 5xx: form stays as-is, equipment choice is preserved, CTA
      // returns to normal (enabled) state — matches AuthScreen's transient formError pattern.
      showSnack(submitErrorMessage(err))
      setSubmitting(false)
    }
  }

  return {
    slotScreen,
    equipmentChoice,
    setEquipmentChoice,
    submitting,
    snack,
    booking,
    retryLoad: loadSlot,
    submit: () => void submit(),
    dismissSuccess: () => navigate('/bookings'),
  }
}
