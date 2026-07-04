import { useEffect, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { bookingsApi } from '../../api/bookingsApi'
import {
  ApiRequestError,
  NetworkError,
  GENERIC_NETWORK_ERROR,
  GENERIC_SERVER_ERROR,
  GENERIC_LOAD_NETWORK_ERROR,
} from '../../core/errors'

interface RatingContext {
  programTitle: string
  startAt: string
  chefName: string
}

type RatingScreenState =
  | { kind: 'loading' }
  | { kind: 'form'; context: RatingContext }
  | { kind: 'unavailable' } // UC-4 E1
  | { kind: 'already_rated' } // UC-4 E2
  | { kind: 'error'; message: string }

function loadErrorMessage(err: unknown): string {
  if (err instanceof NetworkError) return GENERIC_LOAD_NETWORK_ERROR
  if (err instanceof ApiRequestError && err.status >= 500) return GENERIC_SERVER_ERROR
  return GENERIC_LOAD_NETWORK_ERROR
}

function navContextFrom(state: unknown): RatingContext | null {
  const s = state as Partial<RatingContext> | null
  if (s?.programTitle && s?.startAt && s?.chefName) {
    return { programTitle: s.programTitle, startAt: s.startAt, chefName: s.chefName }
  }
  return null
}

// SCR-007 + LOGIC-005. Main flow: SCR-006 passes programTitle/startAt/chefName via navigation
// state, so no request is needed before rendering the form. Reserve flow (direct/refreshed URL,
// e.g. after a PWA reload): falls back to `getBooking` to recover context and to pre-check the
// two guard conditions (already rated / not ratable) before showing the form at all.
export function useRatingScreen() {
  const { bookingId } = useParams<{ bookingId: string }>()
  const navigate = useNavigate()
  const location = useLocation()

  const navContext = navContextFrom(location.state)
  const [screen, setScreen] = useState<RatingScreenState>(navContext ? { kind: 'form', context: navContext } : { kind: 'loading' })
  const [reloadToken, setReloadToken] = useState(0)
  const [rating, setRating] = useState<number | null>(null)
  const [comment, setComment] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [snack, setSnack] = useState<string | null>(null)

  useEffect(() => {
    if (navContext || !bookingId) return
    let cancelled = false
    setScreen({ kind: 'loading' })

    async function loadReserveContext() {
      try {
        const booking = await bookingsApi.getBooking(bookingId!)
        if (cancelled) return
        if (booking.rating != null) {
          setScreen({ kind: 'already_rated' })
          return
        }
        const started = new Date(booking.slot.start_at).getTime() <= Date.now()
        if (booking.status !== 'active' || !started) {
          setScreen({ kind: 'unavailable' })
          return
        }
        setScreen({
          kind: 'form',
          context: {
            programTitle: booking.slot.program.name,
            startAt: booking.slot.start_at,
            chefName: booking.slot.chef.name,
          },
        })
      } catch (err) {
        if (cancelled) return
        if (err instanceof ApiRequestError && err.status === 401) return
        setScreen({ kind: 'error', message: loadErrorMessage(err) })
      }
    }

    void loadReserveContext()
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bookingId, reloadToken])

  function showSnack(text: string) {
    setSnack(text)
    setTimeout(() => setSnack(null), 4000)
  }

  async function submit() {
    if (rating == null || !bookingId || submitting) return
    setSubmitting(true)
    try {
      await bookingsApi.submitRating(bookingId, { stars: rating, comment: comment.trim() === '' ? undefined : comment })
      // AC-001/AC-002: success -> back to SCR-006, which shows "Спасибо за оценку!" as the
      // parent screen (foundations §6.2) and refetches so `rating` is no longer null.
      navigate(`/bookings/${bookingId}`, { state: { justRated: true } })
    } catch (err) {
      if (err instanceof ApiRequestError && err.status === 401) return
      if (err instanceof ApiRequestError && err.status === 409 && err.body.code === 'already_rated') {
        setScreen({ kind: 'already_rated' })
        setSubmitting(false)
        return
      }
      if (err instanceof ApiRequestError && err.status === 422 && err.body.code === 'not_ratable') {
        setScreen({ kind: 'unavailable' })
        setSubmitting(false)
        return
      }
      // UC-4 E3: network/5xx keeps rating/comment on screen, CTA re-enabled for retry.
      showSnack(err instanceof NetworkError ? GENERIC_NETWORK_ERROR : GENERIC_SERVER_ERROR)
      setSubmitting(false)
    }
  }

  return {
    bookingId,
    screen,
    rating,
    comment,
    submitting,
    snack,
    setRating,
    setComment,
    submit: () => void submit(),
    goBack: () => navigate(`/bookings/${bookingId}`),
    retryAfterError: () => setReloadToken((t) => t + 1),
  }
}
