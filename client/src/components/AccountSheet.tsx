import { useState } from 'react'
import { authApi } from '../api/authApi'
import { session } from '../session/session'
import { GENERIC_NETWORK_ERROR } from '../core/errors'

// BS-004. Two steps (§«Элементы экрана»): "Аккаунт" (email read-only + "Выйти из аккаунта") ->
// confirmation ("Выйти"/"Отмена"). No backdrop/swipe close (critical-confirmation exception,
// foundations §4.3) — only the explicit close (✕) on step 1 or "Отмена" on step 2.
export function AccountSheet({ onClose }: { onClose: () => void }) {
  const [step, setStep] = useState<'account' | 'confirm'>('account')
  const [loggingOut, setLoggingOut] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function confirmLogout() {
    setLoggingOut(true)
    setError(null)
    try {
      await authApi.logout()
      session.clear() // App.tsx's session subscription redirects to SCR-001 — no snack needed.
    } catch {
      // BS-004 §«Состояния экрана»: unlike the general LOGIC-005 rule, a logout failure shows
      // inline text inside the sheet (not a snack), and the sheet stays open on step 2.
      setError(GENERIC_NETWORK_ERROR)
      setLoggingOut(false)
    }
  }

  return (
    <div className="sheet-backdrop">
      <div className="sheet">
        {step === 'account' ? (
          <>
            <div className="sheet__header">
              <h2>Аккаунт</h2>
              <button className="sheet__close" aria-label="Закрыть" onClick={onClose}>
                ✕
              </button>
            </div>
            <p className="sheet__email">{session.getEmail()}</p>
            <button className="sheet__action" onClick={() => setStep('confirm')}>
              Выйти из аккаунта
            </button>
          </>
        ) : (
          <>
            <div className="sheet__header">
              <h2>Выйти из аккаунта?</h2>
            </div>
            {error && <div className="form-error">{error}</div>}
            <div className="sheet__actions">
              <button className="sheet__action sheet__action--primary" disabled={loggingOut} onClick={() => setStep('account')}>
                Отмена
              </button>
              <button className="sheet__action sheet__action--danger" disabled={loggingOut} onClick={() => void confirmLogout()}>
                {loggingOut ? '…' : 'Выйти'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
