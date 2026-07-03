package com.chefstol.app.presentation.slots

import com.chefstol.app.core.mvi.ScreenState
import com.chefstol.app.domain.Slot
import kotlinx.datetime.Instant

data class SlotsUiState(
    val screen: ScreenState<List<Slot>> = ScreenState.Loading,
    // null = default period (upcoming 7 days, LOGIC-003); non-null = date_to applied via
    // BS-001. The client never sends date_from — see LOGIC-003.
    val appliedDateTo: Instant? = null,
)

sealed interface SlotsIntent {
    data object LoadInitial : SlotsIntent
    data object Refresh : SlotsIntent
    data object RetryAfterError : SlotsIntent
    data class ApplyDateFilter(val dateTo: Instant) : SlotsIntent
    data object ResetDateFilter : SlotsIntent
}

sealed interface SlotsEffect {
    data class ShowSnackbar(val message: String) : SlotsEffect
}
