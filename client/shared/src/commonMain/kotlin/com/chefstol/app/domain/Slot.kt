package com.chefstol.app.domain

import kotlinx.datetime.Instant

enum class SlotStatus { SCHEDULED, CANCELLED }

// data-model.md → Slot. totalSeats/freeSeats are the source of truth for capacity — the
// 8/12-seat business rule is enforced on the backend when a slot is created, never re-derived
// or hardcoded here (FR-12).
data class Slot(
    val id: String,
    val program: Program,
    val chef: Chef,
    val startAt: Instant,
    val totalSeats: Int,
    val freeSeats: Int,
    val price: Int,
    val status: SlotStatus,
    val cancellationReason: String?,
)
