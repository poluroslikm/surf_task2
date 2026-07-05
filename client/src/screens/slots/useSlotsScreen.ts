import { useEffect, useMemo, useState } from 'react'
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
  // FEAT-09: pure client-side filter over the already-loaded `content` list — no new request,
  // not persisted (plain component state), reset whenever the screen unmounts/remounts.
  const [searchQuery, setSearchQuery] = useState('')

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

  // FEAT-09: case-insensitive substring match against program/chef name, computed from whatever
  // is currently in Content state — never triggers a request (LOGIC-003/listSlots untouched).
  const filteredSlots = useMemo(() => {
    if (screen.kind !== 'content') return []
    const q = searchQuery.trim().toLowerCase()
    if (!q) return screen.value
    return screen.value.filter(
      (slot) => slot.program.name.toLowerCase().includes(q) || slot.chef.name.toLowerCase().includes(q),
    )
  }, [screen, searchQuery])

  return {
    screen,
    appliedDateTo,
    isFilterActive: appliedDateTo !== undefined,
    snack,
    searchQuery,
    setSearchQuery,
    filteredSlots,
    refresh: () => void load(appliedDateTo, true),
    retryAfterError: () => void load(appliedDateTo, false),
    // FEAT-08 (BS-001 "Применить"): sets the filter and re-triggers listSlots with date_to.
    applyDateTo: (date: string) => {
      setAppliedDateTo(date)
      void load(date, false)
    },
    resetDateFilter: () => {
      setAppliedDateTo(undefined)
      void load(undefined, false)
    },
  }
}
