// Machine codes fixed by api/common/models.yaml — mirrors backend/internal/httpapi/errors.go
// exactly. Do not invent new ones without updating the contract first.
export type ErrorCode =
  | 'bad_request'
  | 'unauthorized'
  | 'forbidden'
  | 'not_found'
  | 'internal_error'
  | 'slot_full'
  | 'slot_cancelled'
  | 'slot_started'
  | 'already_cancelled'
  | 'not_ratable'
  | 'already_rated'

export interface ApiError {
  code: ErrorCode
  message: string
  details?: Record<string, unknown>
}

// Sentinel for a request that never reached the server (offline, timeout, CORS, etc.) — has no
// ApiError body to parse.
export class NetworkError extends Error {
  constructor() {
    super('network')
  }
}

// Sentinel for a contractual error response ({code, message}), carrying the HTTP status because
// some cases (register's 409 email-taken) must be told apart by status, not by code — see
// 5-mobile-app-spec/09_Логики/LOGIC-001_Авторизация.md → «API-запросы».
export class ApiRequestError extends Error {
  readonly status: number
  readonly body: ApiError

  constructor(status: number, body: ApiError) {
    super(body.message)
    this.status = status
    this.body = body
  }
}

// Sквозные тексты — 3-design-brief/00-foundations.md §6.
export const GENERIC_NETWORK_ERROR = 'Не удалось выполнить. Проверьте соединение и повторите.'
export const GENERIC_LOAD_NETWORK_ERROR = 'Не удалось загрузить. Проверьте соединение и попробуйте снова.'
export const GENERIC_SERVER_ERROR = 'Что-то пошло не так. Попробуйте ещё раз позже.'
