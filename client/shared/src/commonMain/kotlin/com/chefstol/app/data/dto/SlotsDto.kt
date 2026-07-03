package com.chefstol.app.data.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class ProgramDto(
    val id: String,
    val name: String,
    val difficulty: String,
    @SerialName("photo_url") val photoUrl: String,
    val ingredients: List<String>,
    val allergens: List<String>,
)

@Serializable
data class ChefDto(val id: String, val name: String)

@Serializable
data class SlotDto(
    val id: String,
    val program: ProgramDto,
    val chef: ChefDto,
    @SerialName("start_at") val startAt: String,
    @SerialName("total_seats") val totalSeats: Int,
    @SerialName("free_seats") val freeSeats: Int,
    val price: Int,
    val status: String,
    @SerialName("cancellation_reason") val cancellationReason: String? = null,
)

@Serializable
data class SlotListResponseDto(val items: List<SlotDto>)
