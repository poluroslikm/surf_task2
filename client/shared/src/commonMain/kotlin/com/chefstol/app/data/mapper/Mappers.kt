package com.chefstol.app.data.mapper

import com.chefstol.app.data.dto.ChefDto
import com.chefstol.app.data.dto.ClientDto
import com.chefstol.app.data.dto.ProgramDto
import com.chefstol.app.data.dto.SlotDto
import com.chefstol.app.domain.Chef
import com.chefstol.app.domain.Client
import com.chefstol.app.domain.Program
import com.chefstol.app.domain.ProgramDifficulty
import com.chefstol.app.domain.Slot
import com.chefstol.app.domain.SlotStatus
import kotlinx.datetime.Instant

fun ClientDto.toDomain(): Client = Client(id = id, email = email, createdAt = Instant.parse(createdAt))

fun ProgramDto.toDomain(): Program = Program(
    id = id,
    name = name,
    difficulty = if (difficulty == "experienced") ProgramDifficulty.EXPERIENCED else ProgramDifficulty.NOVICE,
    photoUrl = photoUrl,
    ingredients = ingredients,
    allergens = allergens,
)

fun ChefDto.toDomain(): Chef = Chef(id = id, name = name)

fun SlotDto.toDomain(): Slot = Slot(
    id = id,
    program = program.toDomain(),
    chef = chef.toDomain(),
    startAt = Instant.parse(startAt),
    totalSeats = totalSeats,
    freeSeats = freeSeats,
    price = price,
    status = if (status == "cancelled") SlotStatus.CANCELLED else SlotStatus.SCHEDULED,
    cancellationReason = cancellationReason,
)
