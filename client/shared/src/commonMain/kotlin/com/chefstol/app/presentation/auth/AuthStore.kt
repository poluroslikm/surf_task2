package com.chefstol.app.presentation.auth

import com.chefstol.app.core.error.AppFailure
import com.chefstol.app.core.error.ErrorCodes
import com.chefstol.app.data.repository.AuthOutcome
import com.chefstol.app.data.repository.AuthRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private val EMAIL_REGEX = Regex("^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$")
private const val MIN_PASSWORD_LENGTH = 8

// SCR-001 does not go through ScreenState/LOGIC-005 — it has no GET request on open (see
// SCR-001-auth.md → "Состояния экрана"), so this store keeps its own small local states
// instead of the shared Loading/Content/Empty/Error pattern.
class AuthStore(private val repository: AuthRepository) {

    private val scope = CoroutineScope(SupervisorJob())
    private val _state = MutableStateFlow(AuthUiState())
    val state: StateFlow<AuthUiState> = _state

    private val _effects = Channel<AuthEffect>()
    val effects = _effects.receiveAsFlow()

    fun dispatch(intent: AuthIntent) {
        when (intent) {
            is AuthIntent.TabSelected -> selectTab(intent.tab)
            is AuthIntent.EmailChanged -> _state.update { it.copy(email = intent.value, formError = null) }
            is AuthIntent.PasswordChanged -> _state.update { it.copy(password = intent.value, formError = null) }
            AuthIntent.PasswordVisibilityToggled -> _state.update { it.copy(passwordVisible = !it.passwordVisible) }
            AuthIntent.EmailBlurred -> validateEmail()
            AuthIntent.PasswordBlurred -> validatePasswordIfRegisterTab()
            AuthIntent.Submit -> submit()
        }
    }

    private fun selectTab(tab: AuthTab) {
        // SCR-001 §2 / AC-003: switching tabs clears password but keeps email; the E2
        // password-length check only applies on the Register tab.
        _state.update { it.copy(activeTab = tab, password = "", passwordError = null, formError = null) }
    }

    private fun validateEmail() {
        val email = _state.value.email
        val error = if (email.isNotBlank() && !EMAIL_REGEX.matches(email)) "Введите корректный email" else null
        _state.update { it.copy(emailError = error) }
    }

    private fun validatePasswordIfRegisterTab() {
        val s = _state.value
        if (s.activeTab != AuthTab.REGISTER) return
        val error = if (s.password.isNotEmpty() && s.password.length < MIN_PASSWORD_LENGTH) {
            "Пароль должен быть не короче 8 символов"
        } else {
            null
        }
        _state.update { it.copy(passwordError = error) }
    }

    private fun submit() {
        val s = _state.value
        if (!s.canSubmit) return
        if (!EMAIL_REGEX.matches(s.email)) {
            _state.update { it.copy(emailError = "Введите корректный email") }
            return
        }
        // AC-E02/AC-E03: the 8-char check only blocks submission on the Register tab; on Login
        // it's not applied client-side at all — the server's response decides (401 → E3).
        if (s.activeTab == AuthTab.REGISTER && s.password.length < MIN_PASSWORD_LENGTH) {
            _state.update { it.copy(passwordError = "Пароль должен быть не короче 8 символов") }
            return
        }

        _state.update { it.copy(submitting = true, formError = null) }
        scope.launch {
            val outcome = if (s.activeTab == AuthTab.REGISTER) {
                repository.register(s.email, s.password)
            } else {
                repository.login(s.email, s.password)
            }
            when (outcome) {
                is AuthOutcome.Success -> {
                    _state.update { it.copy(submitting = false) }
                    _effects.send(AuthEffect.NavigateToSlots)
                }
                is AuthOutcome.Failed -> applyFailure(outcome.failure)
            }
        }
    }

    private fun applyFailure(failure: AppFailure) {
        val message = when (failure) {
            is AppFailure.Api -> when {
                // LOGIC-001 → "API-запросы": email-taken is distinguished by HTTP 409, not by
                // Error.code (both are "bad_request").
                failure.httpStatus == 409 -> "Этот email уже зарегистрирован. Войдите или используйте другой email."
                failure.code == ErrorCodes.UNAUTHORIZED -> "Неверный email или пароль"
                else -> failure.message
            }
            AppFailure.Unauthorized -> "Неверный email или пароль"
            AppFailure.NetworkUnavailable -> "Не удалось выполнить. Проверьте соединение и повторите."
            AppFailure.Unknown -> "Что-то пошло не так. Попробуйте ещё раз позже."
        }
        // AC-N03: wrong login credentials clear only the password, keep email, and never
        // reveal which field was wrong.
        val clearPassword = _state.value.activeTab == AuthTab.LOGIN &&
            failure is AppFailure.Api && failure.code == ErrorCodes.UNAUTHORIZED
        _state.update {
            it.copy(
                submitting = false,
                formError = message,
                password = if (clearPassword) "" else it.password,
            )
        }
    }
}
