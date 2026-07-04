import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMyBookingsScreen } from './useMyBookingsScreen'
import type { BookingSummary, EquipmentChoice } from '../../api/types'
import { AlertPotIllustration, EmptyBowlIllustration, WaveDivider } from '../../components/Illustrations'
import { AccountSheet } from '../../components/AccountSheet'
import { TabBar } from '../../components/TabBar'

function equipmentLabel(choice: EquipmentChoice): string {
  return choice === 'own' ? 'Своя экипировка' : 'Прокатная экипировка'
}

// SCR-005 §"Логика статуса и группировки" / SCR-006 §2 — same 4-value badge vocabulary as
// BookingDetailsScreen; `late_cancel` shares the "Отменена" label with `cancelled` (no separate
// permanent status — the early/late distinction only ever shows up as a transient snack).
function bookingBadge(booking: BookingSummary, now: Date): { text: string; className: string } {
  if (booking.status === 'cancelled_by_studio') return { text: 'Отменён студией', className: 'badge--cancelled' }
  if (booking.status === 'cancelled' || booking.status === 'late_cancel') return { text: 'Отменена', className: 'badge--cancelled' }
  const isFuture = new Date(booking.slot.start_at).getTime() > now.getTime()
  return isFuture ? { text: 'Активна', className: 'badge--active' } : { text: 'Прошла', className: 'badge--past' }
}

function BookingCard({ booking, now, onClick }: { booking: BookingSummary; now: Date; onClick: () => void }) {
  const badge = bookingBadge(booking, now)
  return (
    <button className="slot-card" onClick={onClick}>
      <div className="slot-card__date">{new Date(booking.slot.start_at).toLocaleString('ru-RU')}</div>
      <div className="slot-card__title">{booking.slot.program.name}</div>
      <div className="slot-card__chef">Шеф {booking.slot.chef.name}</div>
      <div className="booking-card__row">
        <span className="booking-card__equipment">{equipmentLabel(booking.equipment_choice)}</span>
        <span className="booking-card__price">{booking.price_total} ₽</span>
      </div>
      <div className={`badge ${badge.className}`}>{badge.text}</div>
    </button>
  )
}

function LoadingSkeleton() {
  return (
    <div className="skeleton-list" role="status" aria-label="Загрузка списка бронирований">
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="skeleton-card" aria-hidden="true" />
      ))}
    </div>
  )
}

// SCR-005. Header (title + account icon) and TabBar mirror SlotsScreen (SCR-002) exactly, per
// the task brief. Segmented switcher reuses the .tabs/.tab styling already established on
// SCR-001; cards reuse .slot-card chrome with booking-specific inner rows.
export function MyBookingsScreen() {
  const s = useMyBookingsScreen()
  const navigate = useNavigate()
  const [accountOpen, setAccountOpen] = useState(false)
  const now = new Date()

  const activeList = s.group === 'upcoming' ? s.upcoming : s.past

  return (
    <div className="screen screen--bookings">
      <header className="slots-header">
        <div className="slots-header__row">
          <h1>Мои записи</h1>
          <div className="slots-header__icons">
            <button title="Аккаунт" onClick={() => setAccountOpen(true)}>
              👤
            </button>
          </div>
        </div>
        <WaveDivider color="var(--bg)" />
      </header>

      <div className="screen-scroll">
        {s.screen.kind === 'loading' && <LoadingSkeleton />}

        {s.screen.kind === 'content' && (
          <>
            <div className="tabs" role="tablist">
              <button
                role="tab"
                aria-selected={s.group === 'upcoming'}
                className={`tab ${s.group === 'upcoming' ? 'tab--active' : ''}`}
                onClick={() => s.setGroup('upcoming')}
              >
                Предстоящие
              </button>
              <button
                role="tab"
                aria-selected={s.group === 'past'}
                className={`tab ${s.group === 'past' ? 'tab--active' : ''}`}
                onClick={() => s.setGroup('past')}
              >
                Прошедшие
              </button>
            </div>

            <button className="refresh-button" onClick={s.refresh}>
              {s.screen.refreshing ? 'Обновление…' : 'Обновить'}
            </button>

            {activeList.length === 0 ? (
              <div className="empty-state">
                <EmptyBowlIllustration />
                <p>{s.group === 'upcoming' ? 'Нет предстоящих записей' : 'Нет прошедших записей'}</p>
                {/* AC-E02: the "record a class" CTA is only specified for the empty-upcoming case. */}
                {s.group === 'upcoming' && <button onClick={() => navigate('/')}>Записаться на класс</button>}
              </div>
            ) : (
              <div className="slot-list">
                {activeList.map((booking) => (
                  <BookingCard
                    key={booking.id}
                    booking={booking}
                    now={now}
                    onClick={() => navigate(`/bookings/${booking.id}`)}
                  />
                ))}
              </div>
            )}
          </>
        )}

        {s.screen.kind === 'empty' && (
          <div className="empty-state">
            <EmptyBowlIllustration />
            <p>{s.screen.reason}</p>
            <button onClick={() => navigate('/')}>Записаться на класс</button>
          </div>
        )}

        {s.screen.kind === 'error' && (
          <div className="error-state">
            <AlertPotIllustration />
            <p>{s.screen.message}</p>
            <button onClick={s.retryAfterError}>Обновить</button>
          </div>
        )}

        {s.snack && <div className="snack">{s.snack}</div>}
      </div>

      <TabBar />
      {accountOpen && <AccountSheet onClose={() => setAccountOpen(false)} />}
    </div>
  )
}
