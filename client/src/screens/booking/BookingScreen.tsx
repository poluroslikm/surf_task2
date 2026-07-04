import { useNavigate, useParams } from 'react-router-dom'
import { useBookingScreen } from './useBookingScreen'
import type { ProgramDifficulty } from '../../api/types'
import { AlertPotIllustration, WaveDivider } from '../../components/Illustrations'
import { BookingSuccessSheet } from '../../components/BookingSuccessSheet'

function difficultyLabel(difficulty: ProgramDifficulty): string {
  return difficulty === 'novice' ? 'Новичковый' : 'Для опытных'
}

function difficultyEmoji(difficulty: ProgramDifficulty): string {
  return difficulty === 'novice' ? '🌱' : '🔥'
}

// SCR-004. The only input is equipment_choice (own/rental) — no seat counter, no per-seat
// equipment selection (one booking = one seat, a whole-project assumption). Price is always
// slot.price shown as-is, never recomputed for either equipment option (see SCR-004-booking.md
// "Входные данные" — the rental tariff is an explicit open contract question, not implemented).
export function BookingScreen() {
  const { slotId = '' } = useParams<{ slotId: string }>()
  const navigate = useNavigate()
  const s = useBookingScreen(slotId)

  const slot = s.slotScreen.kind === 'content' ? s.slotScreen.value : undefined

  return (
    <div className="screen screen--booking">
      <header className="slot-card-header">
        <div className="slot-card-header__row">
          <button
            className="icon-button"
            aria-label="Назад"
            disabled={s.submitting}
            onClick={() => navigate(`/slots/${slotId}`)}
          >
            ←
          </button>
          <h1>Оформление записи</h1>
        </div>
        <WaveDivider color="var(--bg)" />
      </header>

      <div className="screen-scroll">
        {s.slotScreen.kind === 'loading' && (
          <div role="status" aria-label="Загрузка данных класса">
            <div className="skeleton-line skeleton-line--title" aria-hidden="true" />
            <div className="skeleton-line" aria-hidden="true" />
            <div className="skeleton-line" aria-hidden="true" />
          </div>
        )}

        {s.slotScreen.kind === 'error' && (
          <div className="error-state">
            <AlertPotIllustration />
            <p>{s.slotScreen.message}</p>
            <button onClick={s.retryLoad}>Обновить</button>
          </div>
        )}

        {slot && (
          <>
            <div className="booking-summary">
              <div className="booking-summary__row">{new Date(slot.start_at).toLocaleString('ru-RU')}</div>
              <div className="booking-summary__row booking-summary__program">
                {slot.program.name} · {difficultyEmoji(slot.program.difficulty)} {difficultyLabel(slot.program.difficulty)}
              </div>
              <div className="booking-summary__row">Шеф {slot.chef.name}</div>
            </div>

            <div className="equipment-toggle" role="group" aria-label="Экипировка">
              <button
                type="button"
                className={`equipment-toggle__option ${
                  s.equipmentChoice === 'own' ? 'equipment-toggle__option--active' : ''
                }`}
                disabled={s.submitting}
                aria-pressed={s.equipmentChoice === 'own'}
                onClick={() => s.setEquipmentChoice('own')}
              >
                Своя экипировка
              </button>
              <button
                type="button"
                className={`equipment-toggle__option ${
                  s.equipmentChoice === 'rental' ? 'equipment-toggle__option--active' : ''
                }`}
                disabled={s.submitting}
                aria-pressed={s.equipmentChoice === 'rental'}
                onClick={() => s.setEquipmentChoice('rental')}
              >
                Прокатная экипировка
              </button>
            </div>
            <p className="equipment-toggle__hint">Прокатный набор выдаётся без ограничений — доступен всегда.</p>

            <div className="slot-card-screen__block">
              <div className="slot-card-screen__value slot-card-screen__price">{slot.price} ₽</div>
              <p className="slot-card-screen__offline-note">Оплата на месте: наличные или перевод на карту.</p>
            </div>
          </>
        )}

        {s.snack && <div className="snack">{s.snack}</div>}
      </div>

      {slot && (
        <div className="bottom-cta">
          <button
            disabled={s.submitting}
            aria-busy={s.submitting}
            aria-label={`Записаться на класс «${slot.program.name}», ${slot.price} рублей`}
            onClick={s.submit}
          >
            {s.submitting ? '…' : `Записаться · ${slot.price} ₽`}
          </button>
        </div>
      )}

      {s.booking && <BookingSuccessSheet booking={s.booking} onDone={s.dismissSuccess} />}
    </div>
  )
}
