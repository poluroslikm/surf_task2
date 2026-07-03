package com.chefstol.app.core.network

import com.chefstol.app.core.error.AppFailure
import com.chefstol.app.data.dto.ErrorDto
import io.ktor.client.call.body
import io.ktor.client.plugins.ResponseException
import io.ktor.client.statement.HttpResponse
import io.ktor.http.HttpStatusCode
import io.ktor.http.isSuccess

sealed interface ApiResult<out T> {
    data class Success<T>(val value: T) : ApiResult<T>
    data class Failure(val failure: AppFailure) : ApiResult<Nothing>
}

suspend inline fun <reified T> safeApiCall(block: suspend () -> HttpResponse): ApiResult<T> {
    return try {
        val response = block()
        if (response.status.isSuccess()) {
            ApiResult.Success(response.body())
        } else {
            ApiResult.Failure(response.toAppFailure())
        }
    } catch (e: ResponseException) {
        ApiResult.Failure(e.response.toAppFailure())
    } catch (e: Exception) {
        ApiResult.Failure(AppFailure.NetworkUnavailable)
    }
}

suspend fun HttpResponse.toAppFailure(): AppFailure {
    if (status == HttpStatusCode.Unauthorized) return AppFailure.Unauthorized
    val error = try {
        body<ErrorDto>()
    } catch (e: Exception) {
        null
    }
    return if (error != null) {
        AppFailure.Api(httpStatus = status.value, code = error.code, message = error.message)
    } else {
        AppFailure.Unknown
    }
}
