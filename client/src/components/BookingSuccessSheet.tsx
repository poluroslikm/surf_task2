import { useEffect, useState } from 'react'
import { pushApi } from '../api/pushApi'
import type { Booking, EquipmentChoice } from '../api/types'
import { VAPID_PUBLIC_KEY } from '../core/config'
import { urlBase64ToUint8Array } from '../core/webPush'

// LOGIC-004: one-time-per-device flag, deliberately not tied to the server profile (§"Пояснения
// к шагам" п.2 — the offer is scoped to this browser/device, not synced across devices).
const PUSH_OFFERED_KEY = 'chef-stol.push_offered'

function equipmentLabel(choice: EquipmentChoice): string {
  return choice === 'own' ? 'Своя экипировка' : 'Прокатная экипировка'
}

// BS-002. Per the spec's own header note, the final implementation is a full-screen step, not a
// swipe/backdrop-dismissable bottom sheet — the only way out is the explicit "К моим записям"
// button. Still reuses the shared .sheet-backdrop/.sheet classes (see AccountSheet.tsx) rather
// than inventing new modal styling; like AccountSheet's own confirm step, the backdrop here has
// no onClick-to-close handler (critical/terminal-step exception, foundations §4.3).
export function BookingSuccessSheet({ booking, onDone }: { booking: Booking; onDone: () => void }) {
  const [showPushBlock, setShowPushBlock] = useState(false)

  useEffect(() => {
    // LOGIC-004 flow, in order: (1) feature-detect Web Push support — unsupported browsers skip
    // silently, no flag write, no error surfaced (best-effort; this is a PWA, not a native SDK).
    const supported = 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window
    if (!supported) return

    // (2) never ask twice on this device.
    if (localStorage.getItem(PUSH_OFFERED_KEY) === 'true') return

    // (3) if the browser already has a decision (granted/denied via browser settings, outside
    // this app), the system dialog can't be re-triggered anyway — just sync the local flag so we
    // stop trying, without showing the block this time.
    if (Notification.permission !== 'default') {
      localStorage.setItem(PUSH_OFFERED_KEY, 'true')
      return
    }

    // (4) Show the block. AC-005: it counts as "shown" the instant it appears — swipe/backdrop/
    // "К моим записям" are all equivalent dismissals, so the flag is set right away rather than
    // only on an explicit "Включить напоминания" tap.
    localStorage.setItem(PUSH_OFFERED_KEY, 'true')
    setShowPushBlock(true)
  }, [])

  function handleEnableReminders() {
    // Best-effort and not awaited — never blocks "К моим записям" (LOGIC-004 "Обработка ошибок":
    // a network failure registering the subscription doesn't block closing BS-002;
    // push_permission = granted is saved locally regardless of server-side outcome). Every step
    // after requestPermission is wrapped so a thrown/rejected promise here never surfaces to the
    // user — at most a console.warn for our own debugging.
    Notification.requestPermission()
      .then(async (permission) => {
        console.log('[push] permission:', permission)
        if (permission !== 'granted') return

        try {
          const registration = await navigator.serviceWorker.ready
          const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
          })
          const json = subscription.toJSON()
          await pushApi.register({ endpoint: json.endpoint!, keys: { p256dh: json.keys!.p256dh, auth: json.keys!.auth } })
        } catch (err) {
          // best-effort — subscribe()/register() failures never block or surface to the user.
          console.warn('[push] subscribe/register failed:', err)
        }
      })
      .catch(() => {
        /* best-effort — ignore */
      })
  }

  return (
    <div className="sheet-backdrop">
      <div className="sheet booking-success-sheet">
        <div className="booking-success-sheet__icon" aria-hidden="true">
          ✅
        </div>
        <h2 className="booking-success-sheet__title">Вы записаны</h2>

        <div className="booking-success-sheet__summary">
          <div className="booking-success-sheet__row">
            <span className="booking-success-sheet__label">Программа</span>
            <span className="booking-success-sheet__value">{booking.slot.program.name}</span>
          </div>
          <div className="booking-success-sheet__row">
            <span className="booking-success-sheet__label">Дата и время</span>
            <span className="booking-success-sheet__value">
              {new Date(booking.slot.start_at).toLocaleString('ru-RU')}
            </span>
          </div>
          <div className="booking-success-sheet__row">
            <span className="booking-success-sheet__label">Шеф</span>
            <span className="booking-success-sheet__value">{booking.slot.chef.name}</span>
          </div>
          <div className="booking-success-sheet__row">
            <span className="booking-success-sheet__label">Экипировка</span>
            <span className="booking-success-sheet__value">{equipmentLabel(booking.equipment_choice)}</span>
          </div>
          <div className="booking-success-sheet__row">
            <span className="booking-success-sheet__label">Цена</span>
            <span className="booking-success-sheet__value booking-success-sheet__price">
              {booking.price_total} ₽
            </span>
          </div>
        </div>

        <p className="booking-success-sheet__offline-note">Оплата на месте: наличные или перевод на карту.</p>

        {showPushBlock && (
          <div className="booking-success-sheet__push">
            <p>Напомним за 24 часа до начала</p>
            <button type="button" className="sheet__action" onClick={handleEnableReminders}>
              Включить напоминания
            </button>
          </div>
        )}

        <button
          type="button"
          className="sheet__action sheet__action--primary booking-success-sheet__done"
          onClick={onDone}
        >
          К моим записям
        </button>
      </div>
    </div>
  )
}
