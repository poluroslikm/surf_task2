import { API_BASE_URL } from '../core/config'
import { session } from '../session/session'
import { ApiRequestError, NetworkError, type ApiError } from '../core/errors'

async function request<T>(path: string, method: 'GET' | 'POST', body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = session.getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  } catch {
    throw new NetworkError()
  }

  if (response.status === 204) {
    return undefined as T
  }

  if (!response.ok) {
    const errorBody = (await response.json()) as ApiError
    // LOGIC-001 AC-007: a 401 on any request other than login itself invalidates the local
    // session — login's own 401 (wrong credentials) is handled by the auth screen, not here.
    if (response.status === 401 && path !== '/auth/login') {
      session.clear()
    }
    throw new ApiRequestError(response.status, errorBody)
  }

  return (await response.json()) as T
}

export const httpClient = {
  get: <T>(path: string) => request<T>(path, 'GET'),
  post: <T>(path: string, body?: unknown) => request<T>(path, 'POST', body),
}
