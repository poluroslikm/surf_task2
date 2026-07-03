package com.chefstol.app.core.di

import com.chefstol.app.data.remote.AuthApi
import com.chefstol.app.data.remote.SlotsApi
import com.chefstol.app.data.repository.AuthRepository
import com.chefstol.app.data.repository.SlotsRepository
import com.chefstol.app.presentation.auth.AuthStore
import com.chefstol.app.presentation.slots.SlotsStore
import org.koin.dsl.module

// Platform-specific pieces (HttpClient engine, SessionStorage implementation) come from
// platformModule (see core/di in the wasmJsMain source set) and must be loaded together with
// this module — see webApp/.../main.kt.
val appModule = module {
    single { AuthApi(get(), get()) }
    single { SlotsApi(get(), get()) }
    single { AuthRepository(get(), get()) }
    single { SlotsRepository(get(), get()) }
    factory { AuthStore(get()) }
    factory { SlotsStore(get()) }
}
