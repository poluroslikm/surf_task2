package com.chefstol.app.domain

// Read-only catalog entry (data-model.md → Chef). Regular and guest chefs are the same
// resource, no separate status (02-domain.md).
data class Chef(val id: String, val name: String)
