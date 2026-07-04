import { useCallback, useEffect, useState } from 'react'
import { slotsApi } from '../../api/slotsApi'
import { ScreenState } from '../../core/screenState'
import { ApiRequestError, NetworkError, GENERIC_LOAD_NETWORK_ERROR, GENERIC_SERVER_ERROR } from '../../core/errors'
import type { Slot } from '../../api/types'

function loadErrorMessage(err: unknown): string {
  if (err instanceof NetworkError) return GENERIC_LOAD_NETWORK_ERROR
  if (err instanceof ApiRequestError && err.status >= 500) return GENERIC_SERVER_ERROR
  return GENERIC_LOAD_NETWORK_ERROR
}

// SCR-003 + LOGIC-005. Only Loading/Content/Error apply here — Empty is not applicable (the
// screen always targets one already-chosen slot, see SCR-003-slot-card.md "Состояния экрана").
// The slot is always re-fetched fresh via getSlot (FE-07/FE-08) — never reused from SCR-002's
// list — so a stale free_seats/status can never be shown right before the client books.
export function useSlotCardScreen(slotId: string) {
  const [screen, setScreen] = useState<ScreenState<Slot>>(ScreenState.loading())

  const load = useCallback(() => {
    setScreen(ScreenState.loading())
    slotsApi
      .getSlot(slotId)
      .then((slot) => setScreen(ScreenState.content(slot)))
      .catch((err) => {
        // Sквозная обработка истёкшей сессии (401 -> logout -> redirect) — happens in
        // httpClient; nothing screen-specific to do here (LOGIC-005).
        if (err instanceof ApiRequestError && err.status === 401) return
        setScreen(ScreenState.error(loadErrorMessage(err)))
      })
  }, [slotId])

  useEffect(() => {
    load()
  }, [load])

  return {
    screen,
    retryAfterError: load,
  }
}
