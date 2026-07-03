package com.chefstol.app.data.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class RegisterRequestDto(val email: String, val password: String)

@Serializable
data class LoginRequestDto(val email: String, val password: String)

@Serializable
data class ClientDto(
    val id: String,
    val email: String,
    @SerialName("created_at") val createdAt: String,
)

@Serializable
data class AuthResponseDto(val token: String, val client: ClientDto)
