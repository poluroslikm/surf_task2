package com.volna.app.map

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.volna.app.core.theme.VolnaTheme
import com.volna.app.domain.model.MeetingPoint
import com.volna.app.domain.model.Route

@Composable
fun RouteMapPreviewFallback(
    route: Route,
    meetingPoint: MeetingPoint,
    onOpenExternal: () -> Unit,
) {
    val spacing = VolnaTheme.tokens.spacing
    val waterColor = Color(0xFF8BD1F1)
    val landColor = Color(0xFFF6F3ED)
    val parkColor = Color(0xFFD5F3BA)
    val streetColor = Color.White
    val roadStrokeColor = Color(0xFFE5DED6)
    val routeColor = Color(0xFF00A59D)
    val pinColor = Color(0xFF00A59D)
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                color = Color(0xFFF2F2F2),
                shape = androidx.compose.foundation.shape.RoundedCornerShape(VolnaTheme.tokens.radius.lg),
            )
            .padding(spacing.md),
        verticalArrangement = Arrangement.spacedBy(spacing.sm),
    ) {
        Text(
            text = "Адрес: ${meetingPoint.title.ifBlank { "уточняется" }}",
            modifier = Modifier.fillMaxWidth(),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurface,
            textAlign = TextAlign.Center,
        )
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(365.dp)
                .clip(androidx.compose.foundation.shape.RoundedCornerShape(VolnaTheme.tokens.radius.md))
                .clickable { onOpenExternal() },
        ) {
            MockRouteScreenshot(
                waterColor = waterColor,
                landColor = landColor,
                parkColor = parkColor,
                streetColor = streetColor,
                roadStrokeColor = roadStrokeColor,
                routeColor = routeColor,
                pinColor = pinColor,
            )
        }
    }
}

@Composable
private fun MockRouteScreenshot(
    waterColor: Color,
    landColor: Color,
    parkColor: Color,
    streetColor: Color,
    roadStrokeColor: Color,
    routeColor: Color,
    pinColor: Color,
) {
    Canvas(
        modifier = Modifier
            .fillMaxWidth()
            .height(365.dp),
    ) {
        val corner = 12.dp.toPx()
        drawRoundRect(
            color = waterColor,
            cornerRadius = CornerRadius(corner, corner),
        )
        drawPath(
            path = androidx.compose.ui.graphics.Path().apply {
                moveTo(0f, 0f)
                lineTo(size.width * 0.16f, 0f)
                cubicTo(size.width * 0.08f, size.height * 0.18f, size.width * 0.18f, size.height * 0.32f, size.width * 0.1f, size.height * 0.48f)
                cubicTo(size.width * 0.02f, size.height * 0.64f, size.width * 0.24f, size.height * 0.78f, size.width * 0.18f, size.height)
                lineTo(0f, size.height)
                close()
            },
            color = landColor,
        )
        drawPath(
            path = androidx.compose.ui.graphics.Path().apply {
                moveTo(0f, size.height * 0.12f)
                lineTo(size.width * 0.06f, size.height * 0.08f)
                lineTo(size.width * 0.04f, size.height * 0.28f)
                lineTo(0f, size.height * 0.33f)
                close()
            },
            color = parkColor,
        )
        drawPath(
            path = androidx.compose.ui.graphics.Path().apply {
                moveTo(0f, size.height * 0.83f)
                cubicTo(size.width * 0.12f, size.height * 0.86f, size.width * 0.25f, size.height * 0.92f, size.width * 0.36f, size.height)
                lineTo(0f, size.height)
                close()
            },
            color = parkColor,
        )
        listOf(0.24f, 0.42f, 0.62f, 0.82f).forEach { y ->
            drawLine(
                color = roadStrokeColor,
                start = Offset(0f, size.height * y),
                end = Offset(size.width * 0.38f, size.height * (y - 0.08f)),
                strokeWidth = 6.dp.toPx(),
                cap = StrokeCap.Round,
            )
            drawLine(
                color = streetColor,
                start = Offset(0f, size.height * y),
                end = Offset(size.width * 0.38f, size.height * (y - 0.08f)),
                strokeWidth = 4.dp.toPx(),
                cap = StrokeCap.Round,
            )
        }
        listOf(0.12f, 0.26f).forEach { x ->
            drawLine(
                color = roadStrokeColor,
                start = Offset(size.width * x, 0f),
                end = Offset(size.width * (x + 0.16f), size.height),
                strokeWidth = 6.dp.toPx(),
                cap = StrokeCap.Round,
            )
            drawLine(
                color = streetColor,
                start = Offset(size.width * x, 0f),
                end = Offset(size.width * (x + 0.16f), size.height),
                strokeWidth = 4.dp.toPx(),
                cap = StrokeCap.Round,
            )
        }
        val routePath = androidx.compose.ui.graphics.Path().apply {
            moveTo(size.width * 0.34f, size.height * 0.57f)
            cubicTo(size.width * 0.58f, size.height * 0.58f, size.width * 0.62f, size.height * 0.48f, size.width * 0.61f, size.height * 0.43f)
            cubicTo(size.width * 0.6f, size.height * 0.36f, size.width * 0.72f, size.height * 0.32f, size.width * 0.82f, size.height * 0.24f)
            lineTo(size.width * 0.82f, 0f)
        }
        drawPath(
            path = routePath,
            color = routeColor,
            style = androidx.compose.ui.graphics.drawscope.Stroke(
                width = 3.dp.toPx(),
                cap = StrokeCap.Round,
            ),
        )
        val pinCenter = Offset(size.width * 0.34f, size.height * 0.57f)
        drawCircle(
            color = routeColor.copy(alpha = 0.16f),
            radius = 32.dp.toPx(),
            center = pinCenter,
        )
        drawCircle(
            color = pinColor,
            radius = 7.dp.toPx(),
            center = pinCenter,
        )
        drawCircle(
            color = Color.White,
            radius = 3.dp.toPx(),
            center = pinCenter,
        )
    }
}
