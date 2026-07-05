import { httpClient } from './httpClient'
import type { PushSubscriptionDto } from './types'

export const pushApi = {
  register: (sub: PushSubscriptionDto) => httpClient.post<PushSubscriptionDto>('/push/subscriptions', sub),
}
