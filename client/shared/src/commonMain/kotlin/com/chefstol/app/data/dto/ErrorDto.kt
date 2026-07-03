package com.chefstol.app.data.dto

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonObject

// api/common/models.yaml#/components/schemas/Error
@Serializable
data class ErrorDto(
    val code: String,
    val message: String,
    val details: JsonObject? = null,
)
