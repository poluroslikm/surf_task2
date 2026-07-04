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
