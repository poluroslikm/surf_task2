import { useRef, useState } from 'react'
import type { TouchEvent } from 'react'

// BS-001. Unlike AccountSheet (BS-004), this sheet is NOT a critical confirmation — backdrop tap,
// swipe-down and ✕ are all allowed close paths, and none of them apply the field's value
// (design-brief §3, BS-001-date-filter.md §"Свойства Bottom Sheet", AC-N01). Only the explicit
// "Применить"/"Сбросить" buttons mutate SCR-002's state.

function todayISO(): string {
  return new Date().toISOString().slice(0, 10)
}

// AC-001/AC-004: pre-fill with the previously applied date_to if a filter is active, otherwise
// with the end of the default 7-day window (today + 7 days).
function defaultDateToISO(): string {
  const d = new Date()
  d.setDate(d.getDate() + 7)
  return d.toISOString().slice(0, 10)
}

export function DateFilterSheet({
  appliedDateTo,
  onClose,
  onApply,
  onReset,
}: {
  appliedDateTo: string | undefined
  onClose: () => void
  onApply: (date: string) => void
  onReset: () => void
}) {
  const isFilterActive = appliedDateTo !== undefined
  const [selectedDate, setSelectedDate] = useState(appliedDateTo ?? defaultDateToISO())
  const touchStartY = useRef<number | null>(null)

  function handleGrabberTouchStart(e: TouchEvent<HTMLDivElement>) {
    touchStartY.current = e.touches[0].clientY
  }

  function handleGrabberTouchEnd(e: TouchEvent<HTMLDivElement>) {
    if (touchStartY.current === null) return
    const delta = e.changedTouches[0].clientY - touchStartY.current
    touchStartY.current = null
    if (delta > 60) onClose() // swipe down closes without applying (AC-N01)
  }

  return (
    <div className="sheet-backdrop" onClick={onClose}>
      <div className="sheet date-filter-sheet" onClick={(e) => e.stopPropagation()}>
        <div
          className="sheet__grabber"
          onTouchStart={handleGrabberTouchStart}
          onTouchEnd={handleGrabberTouchEnd}
        />
        <div className="sheet__header">
          <h2>Фильтр по датам</h2>
          <button className="sheet__close" aria-label="Закрыть" onClick={onClose}>
            ✕
          </button>
        </div>

        <p className="date-filter-sheet__current">
          {isFilterActive
            ? `Сейчас: по ${new Date(appliedDateTo).toLocaleDateString('ru-RU')}`
            : 'Сейчас: ближайшие 7 дней'}
        </p>

        <label className="field">
          <span>Показать по</span>
          <input
            type="date"
            min={todayISO()}
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
          />
        </label>
        <p className="date-filter-sheet__hint">Дата не раньше сегодняшней</p>

        <div className="sheet__actions">
          <button
            className="sheet__action"
            onClick={() => {
              onReset()
              onClose()
            }}
          >
            Сбросить
          </button>
          <button
            className="sheet__action sheet__action--primary"
            onClick={() => {
              onApply(selectedDate)
              onClose()
            }}
          >
            Применить
          </button>
        </div>
      </div>
    </div>
  )
}
