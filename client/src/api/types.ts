// DTOs mirror api/auth/models.yaml and api/slots/models.yaml field-for-field.

export interface RegisterRequest {
  email: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface Client {
  id: string
  email: string
  created_at: string
}

export interface AuthResponse {
  token: string
  client: Client
}

export type ProgramDifficulty = 'novice' | 'experienced'

export interface Program {
  id: string
  name: string
  difficulty: ProgramDifficulty
  photo_url: string
  ingredients: string[]
  allergens: string[]
}

export interface Chef {
  id: string
  name: string
}

export type SlotStatus = 'scheduled' | 'cancelled'

export interface Slot {
  id: string
  program: Program
  chef: Chef
  start_at: string
  total_seats: number
  free_seats: number
  price: number
  status: SlotStatus
  cancellation_reason?: string | null
}

export interface SlotListResponse {
  items: Slot[]
}

export type EquipmentChoice = 'own' | 'rental'

export interface CreateBookingRequest {
  slot_id: string
  equipment_choice: EquipmentChoice
}

export interface SubmitRatingRequest {
  stars: number
  comment?: string | null
}

export interface Rating {
  stars: number
  comment?: string | null
  created_at: string
}

export type BookingStatus = 'active' | 'cancelled' | 'late_cancel' | 'cancelled_by_studio'

export interface Booking {
  id: string
  slot_id: string
  client_id: string
  equipment_choice: EquipmentChoice
  status: BookingStatus
  price_total: number
  free_cancellation_until: string
  rating?: Rating | null
  created_at: string
  cancelled_at?: string | null
  slot: Slot
}

// BookingSummary — deliberately no client_id (per api/bookings/models.yaml).
export interface BookingSummary {
  id: string
  slot_id: string
  equipment_choice: EquipmentChoice
  status: BookingStatus
  price_total: number
  created_at: string
  cancelled_at?: string | null
  slot: Slot
}

export interface BookingListResponse {
  items: BookingSummary[]
}
