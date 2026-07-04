import { httpClient } from './httpClient'
import type { Slot, SlotListResponse } from './types'

export const slotsApi = {
  // LOGIC-003: date_from is never sent by the client — the server defaults it to "now". Only
  // date_to (from BS-001's "Показать по" field) is ever passed explicitly.
  listSlots: (dateTo?: string) => {
    const query = dateTo ? `?date_to=${encodeURIComponent(dateTo)}` : ''
    return httpClient.get<SlotListResponse>(`/slots${query}`)
  },
  getSlot: (slotId: string) => httpClient.get<Slot>(`/slots/${slotId}`),
}
