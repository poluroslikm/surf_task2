package com.chefstol.app.presentation.slots

import com.chefstol.app.core.error.AppFailure
import com.chefstol.app.core.mvi.ScreenState
import com.chefstol.app.data.repository.SlotsOutcome
import com.chefstol.app.data.repository.SlotsRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private const val NETWORK_ERROR_TEXT = "Не удалось загрузить. Проверьте соединение и попробуйте снова."
private const val SERVER_ERROR_TEXT = "Что-то пошло не так. Попробуйте ещё раз позже."
private const val EMPTY_TEXT = "Пока нет доступных классов"

// Implements 09_Логики/LOGIC-005 (Loading/Content/Empty/Error + Refreshing) together with
// LOGIC-003 (date_to-only filtering) for SCR-002.
class SlotsStore(private val repository: SlotsRepository) {

    private val scope = CoroutineScope(SupervisorJob())
    private val _state = MutableStateFlow(SlotsUiState())
    val state: StateFlow<SlotsUiState> = _state

    private val _effects = Channel<SlotsEffect>()
    val effects = _effects.receiveAsFlow()

    fun dispatch(intent: SlotsIntent) {
        when (intent) {
            SlotsIntent.LoadInitial -> load(refreshing = false)
            SlotsIntent.RetryAfterError -> load(refreshing = false)
            SlotsIntent.Refresh -> load(refreshing = true)
            is SlotsIntent.ApplyDateFilter -> {
                _state.update { it.copy(appliedDateTo = intent.dateTo) }
                load(refreshing = false)
            }
            SlotsIntent.ResetDateFilter -> {
                _state.update { it.copy(appliedDateTo = null) }
                load(refreshing = false)
            }
        }
    }

    private fun load(refreshing: Boolean) {
        val current = _state.value.screen

        // LOGIC-005: Refreshing only overlays already-shown Content — it never blanks the
        // screen back to a Loading skeleton, and (handled below) a failure while refreshing
        // never replaces shown data/empty state with an Error screen (AC-006/AC-007).
        if (refreshing && current is ScreenState.Content) {
            _state.update { it.copy(screen = ScreenState.Content(current.value, refreshing = true)) }
        } else if (!refreshing || current !is ScreenState.Empty) {
            _state.update { it.copy(screen = ScreenState.Loading) }
        }

        scope.launch {
            when (val outcome = repository.listSlots(_state.value.appliedDateTo)) {
                is SlotsOutcome.Success -> {
                    _state.update {
                        it.copy(
                            screen = if (outcome.slots.isEmpty()) {
                                ScreenState.Empty(EMPTY_TEXT)
                            } else {
                                ScreenState.Content(outcome.slots)
                            },
                        )
                    }
                }
                is SlotsOutcome.Failed -> {
                    val message = errorText(outcome.failure)
                    when {
                        refreshing && current is ScreenState.Content -> {
                            _state.update { it.copy(screen = ScreenState.Content(current.value, refreshing = false)) }
                            _effects.send(SlotsEffect.ShowSnackbar(message))
                        }
                        refreshing && current is ScreenState.Empty -> {
                            _state.update { it.copy(screen = current) }
                            _effects.send(SlotsEffect.ShowSnackbar(message))
                        }
                        else -> _state.update { it.copy(screen = ScreenState.Error(message)) }
                    }
                }
            }
        }
    }

    private fun errorText(failure: AppFailure): String = when (failure) {
        AppFailure.NetworkUnavailable -> NETWORK_ERROR_TEXT
        else -> SERVER_ERROR_TEXT
    }
}
