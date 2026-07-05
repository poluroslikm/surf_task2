/// <reference lib="webworker" />
import { precacheAndRoute } from 'workbox-precaching'

declare let self: ServiceWorkerGlobalScope

precacheAndRoute(self.__WB_MANIFEST)

// FEAT-07 / LOGIC-004: show a system notification when the backend sends a push message
// (FR-21 24h reminder, FR-22 studio-cancellation notice). Payload shape is backend-defined —
// tolerate a missing/malformed body rather than throwing.
self.addEventListener('push', (event: PushEvent) => {
  const data = (event.data?.json() as { title?: string; body?: string } | undefined) ?? {}
  const title = data.title ?? 'Шеф-стол'
  const body = data.body ?? ''
  event.waitUntil(self.registration.showNotification(title, { body }))
})

self.addEventListener('notificationclick', (event: NotificationEvent) => {
  event.notification.close()
  event.waitUntil(
    self.clients.matchAll({ type: 'window' }).then((clients) => {
      if (clients.length > 0) return (clients[0] as WindowClient).focus()
      return self.clients.openWindow('/')
    }),
  )
})
