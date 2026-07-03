package com.volna.app.core.theme

import androidx.compose.runtime.Immutable
import androidx.compose.ui.graphics.Color

@Immutable
data class VolnaColorScheme(
    val brand: Color,
    val onBrand: Color,
    val background: Color,
    val surface: Color,
    val surfaceVariant: Color,
    val textPrimary: Color,
    val textSecondary: Color,
    val border: Color,
    val success: Color,
    val warning: Color,
    val error: Color,
)

val VolnaLightColors = VolnaColorScheme(
    brand = Color(0xFF00A8A8),
    onBrand = Color.White,
    background = Color.White,
    surface = Color.White,
    surfaceVariant = Color(0xFFF1F1F1),
    textPrimary = Color(0xFF171717),
    textSecondary = Color(0xFF858585),
    border = Color(0xFFE5E5E5),
    success = Color(0xFF237A4B),
    warning = Color(0xFF9A6400),
    error = Color(0xFFB3261E),
)
