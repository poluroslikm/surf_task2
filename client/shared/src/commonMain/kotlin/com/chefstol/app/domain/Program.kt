package com.chefstol.app.domain

enum class ProgramDifficulty { NOVICE, EXPERIENCED }

// Read-only catalog entry (data-model.md → Program). No capacity field here — seat limits
// live on Slot, not on Program (see backend/migrations/00001_init.sql comment: unlike the
// primer/ reference, capacity is not tied to difficulty by a rigid rule).
data class Program(
    val id: String,
    val name: String,
    val difficulty: ProgramDifficulty,
    val photoUrl: String,
    val ingredients: List<String>,
    val allergens: List<String>,
)
