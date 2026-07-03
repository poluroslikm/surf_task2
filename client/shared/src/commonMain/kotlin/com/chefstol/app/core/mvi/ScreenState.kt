package com.chefstol.app.core.mvi

// Direct reflection of 09_Логики/LOGIC-005_Паттерн-состояний-экрана.md — do not add states
// beyond what that document defines (no separate "past"/"upcoming" screen state, etc.).
sealed interface ScreenState<out T> {
    data object Loading : ScreenState<Nothing>
    data class Content<T>(val value: T, val refreshing: Boolean = false) : ScreenState<T>
    data class Empty(val reason: String) : ScreenState<Nothing>
    data class Error(val message: String) : ScreenState<Nothing>
}

// Actions (createBooking/cancelBooking/submitRating in later iterations) never have their own
// Loading/Content/Empty/Error — only a busy flag on the triggering control (LOGIC-005 "Обратная
// связь по действию").
enum class ActionStatus { Idle, Submitting }
