package com.chefstol.app.presentation.slots

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.chefstol.app.core.mvi.ScreenState
import com.chefstol.app.domain.Slot
import com.chefstol.app.domain.SlotStatus
import org.koin.compose.koinInject

// SCR-002. Note: BS-001 (the date-picker bottom sheet UI) is NOT implemented in this pass —
// only the store-level ApplyDateFilter/ResetDateFilter contract exists (see FE report). The
// tab bar from FE-05 is also not wired here since "Мои записи" (SCR-005) doesn't exist yet.
@Composable
fun SlotsScreen(store: SlotsStore = koinInject(), onSlotSelected: (String) -> Unit) {
    val state by store.state.collectAsState()

    LaunchedEffect(Unit) {
        store.dispatch(SlotsIntent.LoadInitial)
    }

    Column(modifier = Modifier.fillMaxSize()) {
        Text("Классы", style = MaterialTheme.typography.headlineSmall, modifier = Modifier.padding(16.dp))

        when (val screen = state.screen) {
            ScreenState.Loading -> LoadingSkeleton()
            is ScreenState.Content -> SlotsList(
                slots = screen.value,
                refreshing = screen.refreshing,
                onSlotClick = onSlotSelected,
                onRefresh = { store.dispatch(SlotsIntent.Refresh) },
            )
            is ScreenState.Empty -> EmptyState(
                message = screen.reason,
                showResetFilter = state.appliedDateTo != null,
                onResetFilter = { store.dispatch(SlotsIntent.ResetDateFilter) },
            )
            is ScreenState.Error -> ErrorState(
                message = screen.message,
                onRetry = { store.dispatch(SlotsIntent.RetryAfterError) },
            )
        }
    }
}

@Composable
private fun LoadingSkeleton() {
    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        repeat(5) {
            Card(modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp)) {
                Text("Загрузка…", modifier = Modifier.padding(16.dp))
            }
        }
    }
}

@Composable
private fun SlotsList(
    slots: List<Slot>,
    refreshing: Boolean,
    onSlotClick: (String) -> Unit,
    onRefresh: () -> Unit,
) {
    Column(modifier = Modifier.fillMaxSize()) {
        // Swipe-to-refresh gesture wiring is deferred (see FE report) — a plain button
        // exercises the same Refresh intent without depending on an unverified
        // pull-to-refresh API surface.
        Button(onClick = onRefresh, modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp)) {
            Text(if (refreshing) "Обновление…" else "Обновить")
        }
        LazyColumn(modifier = Modifier.fillMaxSize().padding(16.dp)) {
            items(slots) { slot -> SlotCard(slot = slot, onClick = { onSlotClick(slot.id) }) }
        }
    }
}

@Composable
private fun SlotCard(slot: Slot, onClick: () -> Unit) {
    Card(onClick = onClick, modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp)) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(slot.startAt.toString(), style = MaterialTheme.typography.bodyMedium)
            Text(slot.program.name, style = MaterialTheme.typography.titleMedium)
            Text(slot.chef.name, style = MaterialTheme.typography.bodySmall)
            // SCR-002 §3: the cancellation badge takes priority over the seats badge — if the
            // studio cancelled the slot, free_seats is never shown, regardless of its value.
            when {
                slot.status == SlotStatus.CANCELLED -> Text("Отменён студией", color = MaterialTheme.colorScheme.error)
                slot.freeSeats == 0 -> Text("Мест нет")
                else -> Text("Свободно: ${slot.freeSeats} из ${slot.totalSeats}")
            }
        }
    }
}

@Composable
private fun EmptyState(message: String, showResetFilter: Boolean, onResetFilter: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(message)
        if (showResetFilter) {
            Button(onClick = onResetFilter, modifier = Modifier.padding(top = 16.dp)) {
                Text("Сбросить фильтр")
            }
        }
    }
}

@Composable
private fun ErrorState(message: String, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(message)
        Button(onClick = onRetry, modifier = Modifier.padding(top = 16.dp)) {
            Text("Обновить")
        }
    }
}
