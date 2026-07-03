package com.chefstol.app.core.storage

import kotlinx.coroutines.flow.Flow

interface SessionStorage {
    suspend fun saveToken(token: String)
    suspend fun currentToken(): String?
    suspend fun clear()
    val isAuthenticated: Flow<Boolean>
}
