package com.chefstol.app.data.repository

import com.chefstol.app.core.error.AppFailure
import com.chefstol.app.core.network.ApiResult
import com.chefstol.app.core.storage.SessionStorage
import com.chefstol.app.data.mapper.toDomain
import com.chefstol.app.data.remote.SlotsApi
import com.chefstol.app.domain.Slot
import kotlinx.datetime.Instant

sealed interface SlotsOutcome {
    data class Success(val slots: List<Slot>) : SlotsOutcome
    data class Failed(val failure: AppFailure) : SlotsOutcome
}

class SlotsRepository(
    private val api: SlotsApi,
    private val session: SessionStorage,
) {
    suspend fun listSlots(dateTo: Instant? = null): SlotsOutcome =
        when (val result = api.listSlots(dateTo)) {
            is ApiResult.Success -> SlotsOutcome.Success(result.value.items.map { it.toDomain() })
            is ApiResult.Failure -> {
                // 09_Логики/LOGIC-001 AC-007: a 401 on any authorized-zone request (other than
                // login itself) must clear the session so the root nav reactively returns to
                // SCR-001 via SessionStorage.isAuthenticated.
                if (result.failure is AppFailure.Unauthorized) session.clear()
                SlotsOutcome.Failed(result.failure)
            }
        }
}
