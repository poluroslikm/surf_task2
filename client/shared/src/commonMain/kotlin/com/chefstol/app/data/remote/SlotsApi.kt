package com.chefstol.app.data.remote

import com.chefstol.app.core.network.ApiResult
import com.chefstol.app.core.network.safeApiCall
import com.chefstol.app.core.storage.SessionStorage
import com.chefstol.app.data.dto.SlotDto
import com.chefstol.app.data.dto.SlotListResponseDto
import io.ktor.client.HttpClient
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.parameter
import io.ktor.http.HttpHeaders
import kotlinx.datetime.Instant

class SlotsApi(
    private val client: HttpClient,
    private val session: SessionStorage,
) {
    // operationId: listSlots. Per 09_Логики/LOGIC-003_Фильтрация-слотов.md the client never
    // sends date_from — only an optional date_to when a filter is applied on BS-001; the
    // server fills in the default 7-day window (FR-3/R-027) when neither is present.
    suspend fun listSlots(dateTo: Instant?): ApiResult<SlotListResponseDto> {
        val token = session.currentToken()
        return safeApiCall {
            client.get("/slots") {
                token?.let { header(HttpHeaders.Authorization, "Bearer $it") }
                dateTo?.let { parameter("date_to", it.toString()) }
            }
        }
    }

    // operationId: getSlot
    suspend fun getSlot(slotId: String): ApiResult<SlotDto> {
        val token = session.currentToken()
        return safeApiCall {
            client.get("/slots/$slotId") {
                token?.let { header(HttpHeaders.Authorization, "Bearer $it") }
            }
        }
    }
}
