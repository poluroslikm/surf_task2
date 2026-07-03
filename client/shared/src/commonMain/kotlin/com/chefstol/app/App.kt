package com.chefstol.app

import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.chefstol.app.core.storage.SessionStorage
import com.chefstol.app.presentation.auth.AuthScreen
import com.chefstol.app.presentation.slots.SlotsScreen
import org.koin.compose.koinInject

// Root switch is intentionally Auth <-> Slots only — no tab bar yet. "Мои записи" (SCR-005)
// does not exist until FE-09, and a tab that goes nowhere would be worse than no tab bar
// (FE-05 in FE_IMPLEMENTATION_PLAN.md is only partially done here, see FE report).
@Composable
fun App(session: SessionStorage = koinInject()) {
    MaterialTheme {
        val isAuthenticated by session.isAuthenticated.collectAsState(initial = false)
        // SCR-003 (slot card) is not implemented yet (FE-07) — selecting a slot is a no-op
        // placeholder, tracked as a known gap in the FE report.
        var selectedSlotId by remember { mutableStateOf<String?>(null) }

        if (!isAuthenticated) {
            // AuthEffect.NavigateToSlots is a no-op here on purpose: the reactive
            // isAuthenticated flow (flipped by SessionStorage.saveToken) is what actually
            // drives this transition; the effect exists for future use (e.g. analytics,
            // one-off UI feedback) and is not currently redundant to remove.
            AuthScreen(onAuthenticated = {})
        } else {
            SlotsScreen(onSlotSelected = { slotId -> selectedSlotId = slotId })
        }
    }
}
