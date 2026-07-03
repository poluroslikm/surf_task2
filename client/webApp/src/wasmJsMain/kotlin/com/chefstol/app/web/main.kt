package com.chefstol.app.web

import androidx.compose.ui.ComposeViewport
import com.chefstol.app.App
import com.chefstol.app.core.di.appModule
import com.chefstol.app.platform.platformModule
import kotlinx.browser.document
import org.koin.core.context.startKoin

fun main() {
    startKoin {
        modules(platformModule, appModule)
    }
    ComposeViewport(document.body!!) {
        App()
    }
}
