package com.chefstol.app.core.storage

import kotlinx.browser.localStorage
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

private const val TOKEN_KEY = "chef_stol_token"

// localStorage, not sessionStorage — 09_Логики/LOGIC-001 "Хранение и использование токена":
// closing the tab/browser normally must not log the client out; only explicit logout or a
// 401 from the server clears the token.
class BrowserSessionStorage : SessionStorage {
    private val state = MutableStateFlow(readToken() != null)

    override suspend fun saveToken(token: String) {
        localStorage.setItem(TOKEN_KEY, token)
        state.value = true
    }

    override suspend fun currentToken(): String? = readToken()

    override suspend fun clear() {
        localStorage.removeItem(TOKEN_KEY)
        state.value = false
    }

    override val isAuthenticated: StateFlow<Boolean> get() = state

    private fun readToken(): String? = localStorage.getItem(TOKEN_KEY)
}
