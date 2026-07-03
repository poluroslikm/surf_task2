package com.chefstol.app.core.error

// httpStatus is kept alongside code because at least one contractual case (register 409)
// must be distinguished by HTTP status, not by Error.code — both are "bad_request" in
// api/common/models.yaml (see 09_Логики/LOGIC-001_Авторизация.md → "API-запросы").
sealed interface AppFailure {
    data object Unauthorized : AppFailure
    data class Api(val httpStatus: Int, val code: String, val message: String) : AppFailure
    data object NetworkUnavailable : AppFailure
    data object Unknown : AppFailure
}

// Mirrors backend/internal/httpapi/errors.go and api/common/models.yaml → Error.code enum.
// Do not invent new codes here without updating the API contract first.
object ErrorCodes {
    const val BAD_REQUEST = "bad_request"
    const val UNAUTHORIZED = "unauthorized"
    const val FORBIDDEN = "forbidden"
    const val NOT_FOUND = "not_found"
    const val SLOT_FULL = "slot_full"
    const val SLOT_CANCELLED = "slot_cancelled"
    const val SLOT_STARTED = "slot_started"
    const val ALREADY_CANCELLED = "already_cancelled"
    const val NOT_RATABLE = "not_ratable"
    const val ALREADY_RATED = "already_rated"
    const val INTERNAL_ERROR = "internal_error"
}
