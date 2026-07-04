import { useNavigate, useParams } from 'react-router-dom'
import { useSlotCardScreen } from './useSlotCardScreen'
import type { ProgramDifficulty } from '../../api/types'
import { AlertPotIllustration, WaveDivider } from '../../components/Illustrations'

function difficultyLabel(difficulty: ProgramDifficulty): string {
  return difficulty === 'novice' ? 'Новичковый' : 'Для опытных'
}

// Purely decorative prefix, mirrors SlotsScreen.tsx — the required label text (checked by AC
// wording) is untouched, this is just a visual cue alongside it (foundations §3.2).
function difficultyEmoji(difficulty: ProgramDifficulty): string {
  return difficulty === 'novice' ? '🌱' : '🔥'
}

function LoadingSkeleton() {
  // Photo-shaped block + a handful of text-line placeholders, per SCR-003 §"Состояния экрана"
  // Loading row (NFR-5) — never a blank screen.
  return (
    <div role="status" aria-label="Загрузка карточки класса">
      <div className="skeleton-photo" aria-hidden="true" />
      <div className="skeleton-line skeleton-line--title" aria-hidden="true" />
      <div className="skeleton-line" aria-hidden="true" />
      <div className="skeleton-line" aria-hidden="true" />
      <div className="skeleton-line" aria-hidden="true" />
    </div>
  )
}

// SCR-003. A pure viewer of getSlot's response (FR-12) — nothing here is editable or
// recomputed client-side. Capacity/status badge logic mirrors SlotsScreen.tsx's SlotCard exactly
// (cancelled takes priority over "no seats", per FR-17/FR-18/FR-6a).
//
// The "Записаться" CTA only renders at all when status='scheduled' && free_seats>0 — no disabled
// button is shown otherwise (per this iteration's decision, matching the list's own pattern of
// never rendering an unusable CTA).
export function SlotCardScreen() {
  const { slotId = '' } = useParams<{ slotId: string }>()
  const navigate = useNavigate()
  const s = useSlotCardScreen(slotId)

  const slot = s.screen.kind === 'content' ? s.screen.value : undefined
  const canBook = !!slot && slot.status === 'scheduled' && slot.free_seats > 0

  return (
    <div className="screen screen--slot-card">
      <header className="slot-card-header">
        <div className="slot-card-header__row">
          <button className="icon-button" aria-label="Назад" onClick={() => navigate('/')}>
            ←
          </button>
          <h1>{slot ? slot.program.name : 'Класс'}</h1>
        </div>
        <WaveDivider color="var(--bg)" />
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

        {slot && (
          <>
            <img
              className="slot-card-screen__photo"
              src={slot.program.photo_url}
              alt={slot.program.name}
            />

            <div className="slot-card-screen__title-row">
              <h2>{slot.program.name}</h2>
              <span className="badge badge--difficulty">
                {difficultyEmoji(slot.program.difficulty)} {difficultyLabel(slot.program.difficulty)}
              </span>
            </div>

            <div className="slot-card-screen__block">
              <div className="slot-card-screen__label">Дата и время</div>
              <div className="slot-card-screen__value slot-card-screen__value--big">
                {new Date(slot.start_at).toLocaleString('ru-RU')}
              </div>
            </div>

            <div className="slot-card-screen__block">
              <div className="slot-card-screen__label">Шеф</div>
              <div className="slot-card-screen__value">{slot.chef.name}</div>
            </div>

            <div className="slot-card-screen__block">
              <div className="slot-card-screen__label">Вместимость</div>
              {slot.status === 'cancelled' ? (
                <div>
                  <div className="badge badge--cancelled">Отменён студией</div>
                  {slot.cancellation_reason && (
                    <p className="slot-card-screen__reason">{slot.cancellation_reason}</p>
                  )}
                </div>
              ) : (
                <div className="slot-card-screen__seats-row">
                  <span className="slot-card-screen__value slot-card-screen__value--big">
                    Всего: {slot.total_seats}
                  </span>
                  {slot.free_seats === 0 ? (
                    <span className="badge badge--full">Мест нет</span>
                  ) : (
                    <span className="slot-card-screen__value slot-card-screen__value--big">
                      Свободно: {slot.free_seats}
                    </span>
                  )}
                </div>
              )}
            </div>

            <div className="slot-card-screen__block">
              <div className="slot-card-screen__label">Состав блюда</div>
              <ul className="slot-card-screen__list">
                {slot.program.ingredients.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            </div>

            {slot.program.allergens.length > 0 && (
              <div className="slot-card-screen__block">
                <div className="slot-card-screen__label">Аллергены</div>
                <ul className="slot-card-screen__list">
                  {slot.program.allergens.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              </div>
            )}

            <div className="slot-card-screen__block">
              <div className="slot-card-screen__value slot-card-screen__price">{slot.price} ₽</div>
              <p className="slot-card-screen__offline-note">Оплата на месте: наличные или перевод на карту.</p>
            </div>
          </>
        )}
      </div>

      {canBook && slot && (
        <div className="bottom-cta">
          <button
            aria-label={`Записаться на класс «${slot.program.name}», ${new Date(slot.start_at).toLocaleDateString('ru-RU')}`}
            onClick={() => navigate(`/slots/${slot.id}/book`)}
          >
            Записаться
          </button>
        </div>
      )}
    </div>
  )
}
