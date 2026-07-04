import { useSlotsScreen } from './useSlotsScreen'
import type { ProgramDifficulty, Slot } from '../../api/types'

function difficultyLabel(difficulty: ProgramDifficulty): string {
  return difficulty === 'novice' ? 'Новичковый' : 'Для опытных'
}

function SlotCard({ slot, onClick }: { slot: Slot; onClick: () => void }) {
  // Field order top-to-bottom per SCR-002-slot-list.md §6.1: date/time -> program + difficulty
  // tag -> chef -> free/total seats or a badge (cancellation badge takes priority over "Мест
  // нет" regardless of free_seats — §6.2/§6.2a).
  return (
    <button className="slot-card" onClick={onClick}>
      <div className="slot-card__date">{new Date(slot.start_at).toLocaleString('ru-RU')}</div>
      <div className="slot-card__title">
        {slot.program.name} · {difficultyLabel(slot.program.difficulty)}
      </div>
      <div className="slot-card__chef">Шеф {slot.chef.name}</div>
      {slot.status === 'cancelled' ? (
        <div className="badge badge--cancelled">Отменён студией</div>
      ) : slot.free_seats === 0 ? (
        <div className="badge badge--full">Мест нет</div>
      ) : (
        <div className="slot-card__seats">
          Свободно: {slot.free_seats} из {slot.total_seats}
        </div>
      )}
    </button>
  )
}

function LoadingSkeleton() {
  // 4-6 card-shaped placeholders, not a blank screen (foundations §5/§7, NFR-5).
  return (
    <div className="skeleton-list">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="skeleton-card">
          Загрузка…
        </div>
      ))}
    </div>
  )
}

// SCR-002. Header/chip/card/tab-bar structure follows the ASCII wireframe in
// 3-design-brief/SCR-002-slot-list.md §5.
//
// Visual placeholders only, wired to no-op (see client/README.md "Не реализовано"):
// - Filter icon (⚲) and "Выбрать другой период" (Empty, default period) — open BS-001, not built.
// - Account icon (⚉) — opens BS-004, not implemented.
// - "Мои записи" tab — SCR-005 doesn't exist yet.
export function SlotsScreen({ onSlotSelected }: { onSlotSelected: (slotId: string) => void }) {
  const s = useSlotsScreen()

  return (
    <div className="screen screen--slots">
      <header className="slots-header">
        <div className="slots-header__row">
          <h1>Классы</h1>
          <div className="slots-header__icons">
            <button title="Фильтр по датам (не реализовано)">🔍</button>
            <button title="Аккаунт (не реализовано)">👤</button>
          </div>
        </div>
        {s.appliedDateTo && (
          <div className="filter-chip">
            {/* Raw ISO string, not the pretty "3-20 июля" format from the wireframe — a Russian
                date formatter isn't wired up in this iteration. */}
            <span>Период: по {s.appliedDateTo}</span>
            <button onClick={s.resetDateFilter}>✕</button>
          </div>
        )}
      </header>

      <div className="screen-scroll">
        {s.screen.kind === 'loading' && <LoadingSkeleton />}

        {s.screen.kind === 'content' && (
          <>
            <button className="refresh-button" onClick={s.refresh}>
              {s.screen.refreshing ? 'Обновление…' : 'Обновить'}
            </button>
            <div className="slot-list">
              {s.screen.value.map((slot) => (
                <SlotCard key={slot.id} slot={slot} onClick={() => onSlotSelected(slot.id)} />
              ))}
            </div>
          </>
        )}

        {s.screen.kind === 'empty' && (
          <div className="empty-state">
            <p>{s.screen.reason}</p>
            {s.appliedDateTo ? (
              <button onClick={s.resetDateFilter}>Сбросить фильтр</button>
            ) : (
              <button title="Не реализовано (BS-001)">Выбрать другой период</button>
            )}
          </div>
        )}

        {s.screen.kind === 'error' && (
          <div className="error-state">
            <p>{s.screen.message}</p>
            <button onClick={s.retryAfterError}>Обновить</button>
          </div>
        )}

        {s.snack && <div className="snack">{s.snack}</div>}
      </div>

      <nav className="tab-bar">
        <span className="tab-bar__item tab-bar__item--active">🍳 Классы</span>
        <span className="tab-bar__item" title="Не реализовано (SCR-005)">
          🗓 Мои записи
        </span>
      </nav>
    </div>
  )
}
