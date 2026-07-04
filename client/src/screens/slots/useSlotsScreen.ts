import { useEffect, useState } from 'react'
import { slotsApi } from '../../api/slotsApi'
import { ScreenState } from '../../core/screenState'
import { ApiRequestError, NetworkError, GENERIC_LOAD_NETWORK_ERROR, GENERIC_SERVER_ERROR } from '../../core/errors'
import type { Slot } from '../../api/types'

function loadErrorMessage(err: unknown): string {
  if (err instanceof NetworkError) return GENERIC_LOAD_NETWORK_ERROR
  if (err instanceof ApiRequestError && err.status >= 500) return GENERIC_SERVER_ERROR
  return GENERIC_LOAD_NETWORK_ERROR
}

// SCR-002 + LOGIC-003 + LOGIC-005.
export function useSlotsScreen() {
  const [screen, setScreen] = useState<ScreenState<Slot[]>>(ScreenState.loading())
  const [appliedDateTo, setAppliedDateTo] = useState<string | undefined>(undefined)
  const [snack, setSnack] = useState<string | null>(null)

  async function load(dateTo: string | undefined, isRefresh: boolean) {
    if (isRefresh) {
      setScreen((prev) => (prev.kind === 'content' ? { ...prev, refreshing: true } : prev))
    } else {
      setScreen(ScreenState.loading())
    }

    try {
      const res = await slotsApi.listSlots(dateTo)
      setScreen(
        res.items.length === 0 ? ScreenState.empty('Пока нет доступных классов') : ScreenState.content(res.items, false),
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
    void load(undefined, false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return {
    screen,
    appliedDateTo,
    snack,
    refresh: () => void load(appliedDateTo, true),
    retryAfterError: () => void load(appliedDateTo, false),
    resetDateFilter: () => {
      setAppliedDateTo(undefined)
      void load(undefined, false)
    },
  }
}
