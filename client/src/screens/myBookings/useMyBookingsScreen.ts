import { useEffect, useState } from 'react'
import { bookingsApi } from '../../api/bookingsApi'
import { ScreenState } from '../../core/screenState'
import { ApiRequestError, NetworkError, GENERIC_LOAD_NETWORK_ERROR, GENERIC_SERVER_ERROR } from '../../core/errors'
import type { BookingSummary } from '../../api/types'

export type BookingListGroup = 'upcoming' | 'past'

function loadErrorMessage(err: unknown): string {
  if (err instanceof NetworkError) return GENERIC_LOAD_NETWORK_ERROR
  if (err instanceof ApiRequestError && err.status >= 500) return GENERIC_SERVER_ERROR
  return GENERIC_LOAD_NETWORK_ERROR
}

// SCR-005 §"Логика статуса и группировки": grouping is by `status` + `slot.start_at`, never by
// `status` alone — a cancelled/late_cancel/cancelled_by_studio booking always lands in
// "Прошедшие" regardless of date, and only a still-`active` booking whose slot hasn't started
// yet counts as "Предстоящие".
export function isUpcoming(booking: BookingSummary, now: Date): boolean {
  return booking.status === 'active' && new Date(booking.slot.start_at).getTime() > now.getTime()
}

// SCR-005 + LOGIC-005.
export function useMyBookingsScreen() {
  const [screen, setScreen] = useState<ScreenState<BookingSummary[]>>(ScreenState.loading())
  const [group, setGroup] = useState<BookingListGroup>('upcoming')
  const [snack, setSnack] = useState<string | null>(null)

  async function load(isRefresh: boolean) {
    if (isRefresh) {
      setScreen((prev) => (prev.kind === 'content' ? { ...prev, refreshing: true } : prev))
    } else {
      setScreen(ScreenState.loading())
    }

    try {
      const res = await bookingsApi.listBookings()
      setScreen(
        res.items.length === 0 ? ScreenState.empty('У вас пока нет записей') : ScreenState.content(res.items, false),
      )
    } catch (err) {
      // Sквозная обработка сессии (401 -> logout -> redirect to SCR-001) happens in httpClient;
      // this screen has nothing screen-specific to do for it (LOGIC-005).
      if (err instanceof ApiRequestError && err.status === 401) return

      if (isRefresh) {
        // LOGIC-005 AC-007: a refresh failure never replaces Content/Empty with Error — just a
        // transient snack, previously shown data stays put.
        setSnack(loadErrorMessage(err))
        setScreen((prev) => (prev.kind === 'content' ? { ...prev, refreshing: false } : prev))
        setTimeout(() => setSnack(null), 4000)
      } else {
        setScreen(ScreenState.error(loadErrorMessage(err)))
      }
    }
  }

  useEffect(() => {
    void load(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const now = new Date()
  let upcoming: BookingSummary[] = []
  let past: BookingSummary[] = []
  if (screen.kind === 'content') {
    // §"Сортировка внутри групп": upcoming re-sorted soonest-first; past keeps the API's
    // descending-by-start_at order as-is (most recent first).
    upcoming = screen.value
      .filter((b) => isUpcoming(b, now))
      .sort((a, b) => new Date(a.slot.start_at).getTime() - new Date(b.slot.start_at).getTime())
    past = screen.value.filter((b) => !isUpcoming(b, now))
  }

  return {
    screen,
    group,
    setGroup,
    upcoming,
    past,
    snack,
    refresh: () => void load(true),
    retryAfterError: () => void load(false),
  }
}
