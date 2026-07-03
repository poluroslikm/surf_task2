package com.chefstol.app.platform

import com.chefstol.app.core.config.AppConfig
import com.chefstol.app.core.storage.BrowserSessionStorage
import com.chefstol.app.core.storage.SessionStorage
import io.ktor.client.HttpClient
import io.ktor.client.engine.js.Js
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.request.url
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json
import org.koin.dsl.module

// Only a wasmJs/web target exists in this iteration (FE plan: web is the priority platform,
// Android/iOS deferred) — so HttpClient construction lives directly here rather than behind a
// commonMain factory abstracting an engine we don't have a second implementation of yet.
val platformModule = module {
    single {
        HttpClient(Js) {
            install(ContentNegotiation) {
                json(Json { ignoreUnknownKeys = true })
            }
            install(HttpTimeout) {
                requestTimeoutMillis = 10_000
            }
            defaultRequest {
                url(AppConfig.API_BASE_URL)
            }
        }
    }
    single<SessionStorage> { BrowserSessionStorage() }
}
