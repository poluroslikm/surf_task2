package com.chefstol.app.presentation.auth

enum class AuthTab { LOGIN, REGISTER }

data class AuthUiState(
    val activeTab: AuthTab = AuthTab.LOGIN,
    val email: String = "",
    val password: String = "",
    val passwordVisible: Boolean = false,
    val emailError: String? = null,
    val passwordError: String? = null,
    val formError: String? = null,
    val submitting: Boolean = false,
) {
    // AC-E01 (SCR-001): CTA disabled while either field is empty — no server round-trip
    // needed to know that.
    val canSubmit: Boolean get() = email.isNotBlank() && password.isNotBlank() && !submitting
}

sealed interface AuthIntent {
    data class TabSelected(val tab: AuthTab) : AuthIntent
    data class EmailChanged(val value: String) : AuthIntent
    data class PasswordChanged(val value: String) : AuthIntent
    data object PasswordVisibilityToggled : AuthIntent
    data object EmailBlurred : AuthIntent
    data object PasswordBlurred : AuthIntent
    data object Submit : AuthIntent
}

sealed interface AuthEffect {
    data object NavigateToSlots : AuthEffect
}
