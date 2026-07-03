package com.chefstol.app.data.repository

import com.chefstol.app.core.error.AppFailure
import com.chefstol.app.core.network.ApiResult
import com.chefstol.app.core.storage.SessionStorage
import com.chefstol.app.data.dto.AuthResponseDto
import com.chefstol.app.data.mapper.toDomain
import com.chefstol.app.data.remote.AuthApi
import com.chefstol.app.domain.Client

sealed interface AuthOutcome {
    data class Success(val client: Client) : AuthOutcome
    data class Failed(val failure: AppFailure) : AuthOutcome
}

class AuthRepository(
    private val api: AuthApi,
    private val session: SessionStorage,
) {
    suspend fun register(email: String, password: String): AuthOutcome = handle(api.register(email, password))

    suspend fun login(email: String, password: String): AuthOutcome = handle(api.login(email, password))

    // Clears the local session regardless of whether the server call succeeds — a client that
    // can't reach the server should still be able to forget its own token.
    suspend fun logout() {
        api.logout()
        session.clear()
    }

    private suspend fun handle(result: ApiResult<AuthResponseDto>): AuthOutcome = when (result) {
        is ApiResult.Success -> {
            session.saveToken(result.value.token)
            AuthOutcome.Success(result.value.client.toDomain())
        }
        is ApiResult.Failure -> AuthOutcome.Failed(result.failure)
    }
}
