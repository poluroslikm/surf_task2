package com.chefstol.app.domain

import kotlinx.datetime.Instant

data class Client(
    val id: String,
    val email: String,
    val createdAt: Instant,
)
