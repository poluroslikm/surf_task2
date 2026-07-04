import { useRatingScreen } from './useRatingScreen'
import { AlertPotIllustration } from '../../components/Illustrations'

function LoadingSkeleton() {
  return (
    <div role="status" aria-label="Загрузка контекста класса" className="skeleton-list">
      <div className="skeleton-card" aria-hidden="true" />
    </div>
  )
}

// SCR-007. Nested screen, tab-bar hidden. Star row/comment/CTA only render in the "form" state —
// the two guard states ("Недоступно" / "Уже оценено") replace the whole form with an explanation
// + "Назад", never partially disable it.
export function RatingScreen() {
  const s = useRatingScreen()

  return (
    <div className="screen screen--rating">
      <header className="details-header">
        <button className="details-header__back" aria-label="Назад" onClick={s.goBack} disabled={s.submitting}>
          ←
        </button>
        <h1>Оценка шефа</h1>
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

        {s.screen.kind === 'unavailable' && (
          <div className="empty-state">
            <p>Оценка недоступна: запись отменена или класс ещё не начался.</p>
            <button onClick={s.goBack}>Назад</button>
          </div>
        )}

        {s.screen.kind === 'already_rated' && (
          <div className="empty-state">
            <p>Вы уже оценили этот класс. Изменить оценку нельзя.</p>
            <button onClick={s.goBack}>Назад</button>
          </div>
        )}

        {s.screen.kind === 'form' && (
          <>
            <section className="detail-section">
              <span className="detail-section__value">{s.screen.context.programTitle}</span>
              <span className="detail-section__hint">{new Date(s.screen.context.startAt).toLocaleString('ru-RU')}</span>
            </section>

            <p className="rating-question">Как вам занятие с шефом {s.screen.context.chefName}?</p>

            <div className="rating-stars" role="radiogroup" aria-label="Оценка от 1 до 5 звёзд">
              {[1, 2, 3, 4, 5].map((n) => (
                <button
                  key={n}
                  type="button"
                  role="radio"
                  aria-checked={s.rating === n}
                  aria-label={`${n} из 5`}
                  className={`rating-star ${s.rating != null && n <= s.rating ? 'rating-star--filled' : ''}`}
                  disabled={s.submitting}
                  onClick={() => s.setRating(n)}
                >
                  {s.rating != null && n <= s.rating ? '★' : '☆'}
                </button>
              ))}
            </div>
            <p className="rating-state-text">{s.rating == null ? 'Оценка: не выбрана' : `Оценка: ${s.rating} из 5`}</p>

            <label className="field">
              <span>Комментарий (необязательно)</span>
              <textarea
                rows={4}
                value={s.comment}
                disabled={s.submitting}
                onChange={(e) => s.setComment(e.target.value)}
              />
            </label>

            {s.snack && <div className="snack">{s.snack}</div>}
          </>
        )}
      </div>

      {s.screen.kind === 'form' && (
        <div className="bottom-cta">
          <button disabled={s.rating == null || s.submitting} onClick={s.submit}>
            {s.submitting ? '…' : 'Отправить оценку'}
          </button>
        </div>
      )}
    </div>
  )
}
