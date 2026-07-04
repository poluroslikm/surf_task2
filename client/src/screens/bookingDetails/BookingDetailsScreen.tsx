import { useNavigate } from 'react-router-dom'
import { useBookingDetailsScreen } from './useBookingDetailsScreen'
import { CancelConfirmSheet } from '../../components/CancelConfirmSheet'
import { AlertPotIllustration } from '../../components/Illustrations'
import type { Booking, EquipmentChoice, ProgramDifficulty } from '../../api/types'

function equipmentLabel(choice: EquipmentChoice): string {
  return choice === 'own' ? 'Своя экипировка' : 'Прокатная экипировка'
}

// SCR-006 §4: this screen's own wording for the difficulty tag — distinct casing from SCR-002's
// list-card label, each screen document specifies its own copy.
function difficultyLabel(difficulty: ProgramDifficulty): string {
  return difficulty === 'novice' ? 'для новичков' : 'для опытных'
}

type UiStatus = 'active' | 'past' | 'cancelled' | 'cancelled_by_studio'

// SCR-006 §2 "Логика — соответствие API-статуса и UI-бейджа". `late_cancel` shares the same
// "Отменена" UI status as `cancelled` — there is no separate permanent late-cancel status.
function uiStatus(booking: Booking, now: Date): UiStatus {
  if (booking.status === 'cancelled_by_studio') return 'cancelled_by_studio'
  if (booking.status === 'cancelled' || booking.status === 'late_cancel') return 'cancelled'
  return new Date(booking.slot.start_at).getTime() > now.getTime() ? 'active' : 'past'
}

const STATUS_BADGE: Record<UiStatus, { text: string; className: string }> = {
  active: { text: 'Активна', className: 'badge--active' },
  past: { text: 'Прошла', className: 'badge--past' },
  cancelled: { text: 'Отменена', className: 'badge--cancelled' },
  cancelled_by_studio: { text: 'Отменён студией', className: 'badge--cancelled' },
}

function LoadingSkeleton() {
  return (
    <div role="status" aria-label="Загрузка деталей записи" className="skeleton-list">
      <div className="skeleton-card" aria-hidden="true" />
      <div className="skeleton-card" aria-hidden="true" />
      <div className="skeleton-card" aria-hidden="true" />
    </div>
  )
}

function RatingDisplay({ stars, comment }: { stars: number; comment?: string | null }) {
  return (
    <div className="rating-display">
      <div className="rating-display__stars" aria-label={`Оценка: ${stars} из 5`}>
        {Array.from({ length: 5 }).map((_, i) => (
          <span key={i} aria-hidden="true">
            {i < stars ? '★' : '☆'}
          </span>
        ))}
      </div>
      {comment && <p className="rating-display__comment">{comment}</p>}
    </div>
  )
}

// SCR-006. Read-only detail screen + cancel entry point (BS-003 as an overlay, foundations §4.3)
// + chef-rating entry point (SCR-007). Bottom-CTA precedence follows the doc's states table
// (§"Состояния экрана") exactly: Активна -> "Отменить" enabled; Прошла + no rating -> "Оценить
// шефа"; Прошла + rating -> read-only stars/comment; Отменена / Отменён студией -> "Отменить"
// disabled with the matching explanation. No separate disabled-cancel state is rendered for
// "Прошла" — the rate CTA / rating display always takes over there.
export function BookingDetailsScreen() {
  const s = useBookingDetailsScreen()
  const navigate = useNavigate()
  const now = new Date()

  return (
    <div className="screen screen--booking-details">
      <header className="details-header">
        <button className="details-header__back" aria-label="Назад" onClick={() => navigate('/bookings')}>
          ←
        </button>
        <h1>Детали записи</h1>
      </header>

      <div className="screen-scroll">
        {s.screen.kind === 'loading' && <LoadingSkeleton />}

        {s.screen.kind === 'error' && (
          <div className="error-state">
            <AlertPotIllustration />
            <p>{s.screen.message}</p>
            <button onClick={s.retryAfterError}>Обновить</button>
          </div>
        )}

        {s.screen.kind === 'content' &&
          (() => {
            const booking = s.screen.value
            const status = uiStatus(booking, now)
            const badge = STATUS_BADGE[status]

            return (
              <>
                <div className="status-badge">
                  <span className={`badge ${badge.className}`}>{badge.text}</span>
                  {status === 'cancelled_by_studio' && (
                    <p className="status-badge__reason">
                      Класс отменён студией. Причина: {booking.slot.cancellation_reason}. Повторная запись на этот
                      слот недоступна.
                    </p>
                  )}
                </div>

                <section className="detail-section">
                  <span className="detail-section__label">Когда</span>
                  <span className="detail-section__value">{new Date(booking.slot.start_at).toLocaleString('ru-RU')}</span>
                </section>

                <section className="detail-section">
                  <span className="detail-section__label">Программа и шеф</span>
                  <span className="detail-section__value">{booking.slot.program.name}</span>
                  <span className="detail-section__hint">{difficultyLabel(booking.slot.program.difficulty)}</span>
                  <span className="detail-section__hint">Шеф {booking.slot.chef.name}</span>
                </section>

                <section className="detail-section">
                  <span className="detail-section__label">Экипировка</span>
                  <span className="detail-section__value">{equipmentLabel(booking.equipment_choice)}</span>
                </section>

                <section className="detail-section">
                  <span className="detail-section__label">Цена</span>
                  <span className="detail-section__value">{booking.price_total} ₽</span>
                  <span className="detail-section__hint">Оплата на месте: наличные или перевод на карту.</span>
                </section>

                {status === 'active' && (
                  <div className="free-cancel-row">
                    Бесплатно отменить можно до {new Date(booking.free_cancellation_until).toLocaleString('ru-RU')}
                  </div>
                )}

                {status === 'past' && booking.rating != null && (
                  <section className="detail-section">
                    <span className="detail-section__label">Ваша оценка</span>
                    <RatingDisplay stars={booking.rating.stars} comment={booking.rating.comment} />
                  </section>
                )}

                {s.snack && <div className="snack">{s.snack}</div>}
              </>
            )
          })()}
      </div>

      {s.screen.kind === 'content' &&
        (() => {
          const booking = s.screen.value
          const status = uiStatus(booking, now)

          if (status === 'active') {
            return (
              <div className="bottom-cta">
                <button onClick={s.openCancelSheet}>Отменить</button>
              </div>
            )
          }
          if (status === 'past' && booking.rating == null) {
            return (
              <div className="bottom-cta">
                <button
                  onClick={() =>
                    navigate(`/bookings/${booking.id}/rate`, {
                      state: {
                        programTitle: booking.slot.program.name,
                        startAt: booking.slot.start_at,
                        chefName: booking.slot.chef.name,
                      },
                    })
                  }
                >
                  Оценить шефа
                </button>
              </div>
            )
          }
          if (status === 'past') {
            return null // Rating already shown inline above (detail-section) — no bottom CTA.
          }
          // cancelled / cancelled_by_studio: disabled "Отменить" with the matching explanation.
          return (
            <div className="bottom-cta">
              <button disabled>Отменить</button>
              <p className="bottom-cta__hint">
                {status === 'cancelled' ? 'Запись уже отменена.' : 'Класс отменён студией — отменять уже нечего.'}
              </p>
            </div>
          )
        })()}

      {s.cancelSheetOpen && s.screen.kind === 'content' && (
        <CancelConfirmSheet
          bookingId={s.screen.value.id}
          freeCancellationUntil={s.screen.value.free_cancellation_until}
          onClose={s.closeCancelSheet}
          onOutcome={s.handleCancelOutcome}
        />
      )}
    </div>
  )
}
