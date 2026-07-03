package com.chefstol.app.data.remote

import com.chefstol.app.core.network.ApiResult
import com.chefstol.app.core.network.safeApiCall
import com.chefstol.app.core.network.toAppFailure
import com.chefstol.app.core.storage.SessionStorage
import com.chefstol.app.data.dto.AuthResponseDto
import com.chefstol.app.data.dto.LoginRequestDto
import com.chefstol.app.data.dto.RegisterRequestDto
import io.ktor.client.HttpClient
import io.ktor.client.request.header
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.contentType
import io.ktor.http.isSuccess

class AuthApi(
    private val client: HttpClient,
    private val session: SessionStorage,
) {
    // operationId: register — no Authorization header: auth/api.yaml only exempts
    // register/login from bearerAuth.
    suspend fun register(email: String, password: String): ApiResult<AuthResponseDto> =
        safeApiCall {
            client.post("/auth/register") {
                contentType(ContentType.Application.Json)
                setBody(RegisterRequestDto(email, password))
            }
        }

    // operationId: login
    suspend fun login(email: String, password: String): ApiResult<AuthResponseDto> =
        safeApiCall {
            client.post("/auth/login") {
                contentType(ContentType.Application.Json)
                setBody(LoginRequestDto(email, password))
            }
        }

    // operationId: logout — 204 No Content, checked by status only rather than relying on
    // ContentNegotiation to decode an empty body as Unit.
    suspend fun logout(): ApiResult<Unit> {
        val token = session.currentToken()
        val response = client.post("/auth/logout") {
            token?.let { header(HttpHeaders.Authorization, "Bearer $it") }
        }
        return if (response.status.isSuccess()) {
            ApiResult.Success(Unit)
        } else {
            ApiResult.Failure(response.toAppFailure())
        }
    }
}
