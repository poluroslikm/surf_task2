import { httpClient } from './httpClient'
import type { AuthResponse, LoginRequest, RegisterRequest } from './types'

export const authApi = {
  register: (req: RegisterRequest) => httpClient.post<AuthResponse>('/auth/register', req),
  login: (req: LoginRequest) => httpClient.post<AuthResponse>('/auth/login', req),
  logout: () => httpClient.post<void>('/auth/logout'),
}
