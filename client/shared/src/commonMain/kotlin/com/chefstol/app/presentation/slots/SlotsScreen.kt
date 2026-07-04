package com.chefstol.app.presentation.slots

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.chefstol.app.core.mvi.ScreenState
import com.chefstol.app.domain.ProgramDifficulty
import com.chefstol.app.domain.Slot
import com.chefstol.app.domain.SlotStatus
import org.koin.compose.koinInject

// SCR-002. Каркас and content order follow the ASCII wireframe in
// 3-design-brief/SCR-002-slot-list.md §5: header (title + filter/account icons) -> optional
// active-filter chip -> card list -> fixed tab bar (foundations §4.1/§4.2).
//
// Visual placeholders only, wired to no-op or a TODO (not real navigation — see
// client/README.md "Не реализовано"):
// - Filter icon (⚲) and "Выбрать другой период" (Empty, default period) — open BS-001, which
//   has no UI yet, only the store-level ApplyDateFilter/ResetDateFilter contract.
// - Account icon (⚉) — opens BS-004, not implemented.
// - "Мои записи" tab — SCR-005 doesn't exist yet (FE-09).
@Composable
fun SlotsScreen(store: SlotsStore = koinInject(), onSlotSelected: (String) -> Unit) {
    val state by store.state.collectAsState()

    LaunchedEffect(Unit) {
        store.dispatch(SlotsIntent.LoadInitial)
    }

    Scaffold(
        topBar = {
            Column {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Text("Классы", style = MaterialTheme.typography.headlineSmall)
                    Row {
                        TextButton(onClick = { /* BS-001 not implemented yet */ }) { Text("⚲") }
                        TextButton(onClick = { /* BS-004 not implemented yet */ }) { Text("⚉") }
                    }
                }
                // SCR-002 §6.3: the chip only appears when the active period differs from the
                // default 7 days, with a quick-reset ✕ (already wired — ResetDateFilter exists).
                if (state.appliedDateTo != null) {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 4.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                    ) {
                        // Raw ISO instant, not the pretty "3-20 июля" format from the wireframe
                        // — a Russian date formatter isn't wired up in this iteration.
                        Text("Период: по ${state.appliedDateTo}", style = MaterialTheme.typography.bodySmall)
                        TextButton(onClick = { store.dispatch(SlotsIntent.ResetDateFilter) }) { Text("✕") }
                    }
                }
            }
        },
        bottomBar = {
            // Tab bar visible across all screen states (Loading/Content/Empty/Error) per
            // foundations §4.2/§6.5 — "Мои записи" doesn't navigate anywhere yet.
            Row(
                modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
                horizontalArrangement = Arrangement.SpaceEvenly,
            ) {
                Text("🍳 Классы", style = MaterialTheme.typography.labelLarge)
                Text(
                    "🗓 Мои записи",
                    style = MaterialTheme.typography.labelLarge,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        },
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
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
}

@Composable
private fun LoadingSkeleton() {
    // 4-6 card-shaped placeholders, not a blank screen (foundations §5/§7, NFR-5).
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
        // Swipe-to-refresh gesture wiring is deferred (see client/README.md) — a plain button
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

private fun difficultyLabel(difficulty: ProgramDifficulty): String = when (difficulty) {
    ProgramDifficulty.NOVICE -> "Новичковый"
    ProgramDifficulty.EXPERIENCED -> "Для опытных"
}

@Composable
private fun SlotCard(slot: Slot, onClick: () -> Unit) {
    // Field order top-to-bottom per SCR-002-slot-list.md §6.1: date/time -> program + difficulty
    // tag -> chef -> free/total seats or a badge (cancellation badge takes priority over
    // "Мест нет" regardless of free_seats — §6.2/§6.2a).
    Card(onClick = onClick, modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp)) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(slot.startAt.toString(), style = MaterialTheme.typography.bodyMedium)
            Text(
                "${slot.program.name} · ${difficultyLabel(slot.program.difficulty)}",
                style = MaterialTheme.typography.titleMedium,
            )
            Text("Шеф ${slot.chef.name}", style = MaterialTheme.typography.bodySmall)
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
            // SCR-002 §8 (E1, custom filter active): already wired to a real intent.
            Button(onClick = onResetFilter, modifier = Modifier.padding(top = 16.dp)) {
                Text("Сбросить фильтр")
            }
        } else {
            // SCR-002 §8 (E1, default period): should open BS-001 — not implemented yet.
            Button(onClick = { /* BS-001 not implemented yet */ }, modifier = Modifier.padding(top = 16.dp)) {
                Text("Выбрать другой период")
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
