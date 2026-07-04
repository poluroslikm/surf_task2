package com.chefstol.app.presentation.auth

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Button
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import org.koin.compose.koinInject

// SCR-001. Каркас follows 3-design-brief/00-foundations.md §4.1 (fixed header without a back
// button — this is the root screen — scroll content, fixed bottom CTA never covered by the
// keyboard) and the ASCII wireframe in 3-design-brief/SCR-001-auth.md §5: title -> tabs ->
// email -> password -> error area -> fixed CTA.
@Composable
fun AuthScreen(store: AuthStore = koinInject(), onAuthenticated: () -> Unit) {
    val state by store.state.collectAsState()

    LaunchedEffect(Unit) {
        store.effects.collect { effect ->
            when (effect) {
                AuthEffect.NavigateToSlots -> onAuthenticated()
            }
        }
    }

    // onFocusChanged fires once on initial composition too (isFocused = false) — guard with
    // a "was it ever focused" flag so validation doesn't fire before the client typed anything.
    var emailFocusedOnce by remember { mutableStateOf(false) }
    var passwordFocusedOnce by remember { mutableStateOf(false) }

    Scaffold(
        modifier = Modifier.imePadding(),
        topBar = {
            Text(
                "Шеф-стол",
                style = MaterialTheme.typography.headlineMedium,
                textAlign = TextAlign.Center,
                modifier = Modifier.fillMaxWidth().padding(24.dp),
            )
        },
        bottomBar = {
            // Fixed bottom CTA (foundations §4.1) — always visible above the keyboard
            // (Scaffold + imePadding), never scrolls away with the form fields.
            Button(
                onClick = { store.dispatch(AuthIntent.Submit) },
                enabled = state.canSubmit,
                modifier = Modifier.fillMaxWidth().padding(16.dp),
            ) {
                if (state.submitting) {
                    CircularProgressIndicator(modifier = Modifier.size(20.dp))
                } else {
                    Text(if (state.activeTab == AuthTab.REGISTER) "Зарегистрироваться" else "Войти")
                }
            }
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 24.dp)
                .verticalScroll(rememberScrollState()),
        ) {
            TabRow(selectedTabIndex = if (state.activeTab == AuthTab.LOGIN) 0 else 1) {
                Tab(
                    selected = state.activeTab == AuthTab.LOGIN,
                    onClick = { store.dispatch(AuthIntent.TabSelected(AuthTab.LOGIN)) },
                    text = { Text("Вход") },
                )
                Tab(
                    selected = state.activeTab == AuthTab.REGISTER,
                    onClick = { store.dispatch(AuthIntent.TabSelected(AuthTab.REGISTER)) },
                    text = { Text("Регистрация") },
                )
            }

            OutlinedTextField(
                value = state.email,
                onValueChange = { store.dispatch(AuthIntent.EmailChanged(it)) },
                label = { Text("Email") },
                singleLine = true,
                enabled = !state.submitting,
                isError = state.emailError != null,
                supportingText = { state.emailError?.let { Text(it) } },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email),
                modifier = Modifier.fillMaxWidth().padding(top = 16.dp).onFocusChanged { focus ->
                    if (focus.isFocused) {
                        emailFocusedOnce = true
                    } else if (emailFocusedOnce) {
                        store.dispatch(AuthIntent.EmailBlurred)
                    }
                },
            )

            OutlinedTextField(
                value = state.password,
                onValueChange = { store.dispatch(AuthIntent.PasswordChanged(it)) },
                label = { Text("Пароль") },
                singleLine = true,
                enabled = !state.submitting,
                isError = state.passwordError != null,
                supportingText = { state.passwordError?.let { Text(it) } },
                visualTransformation = if (state.passwordVisible) VisualTransformation.None else PasswordVisualTransformation(),
                trailingIcon = {
                    // Textual toggle instead of an icon glyph: self-describing for screen
                    // readers without a separate contentDescription/aria-label (foundations §9).
                    IconButton(onClick = { store.dispatch(AuthIntent.PasswordVisibilityToggled) }) {
                        Text(if (state.passwordVisible) "Скрыть" else "Показать")
                    }
                },
                modifier = Modifier.fillMaxWidth().padding(top = 8.dp).onFocusChanged { focus ->
                    if (focus.isFocused) {
                        passwordFocusedOnce = true
                    } else if (passwordFocusedOnce) {
                        store.dispatch(AuthIntent.PasswordBlurred)
                    }
                },
            )

            // Error area — reserves no space in Content state, only appears on E1-E6
            // (SCR-001-auth.md §5, §8).
            state.formError?.let {
                Text(it, color = MaterialTheme.colorScheme.error, modifier = Modifier.padding(top = 8.dp))
            }
        }
    }
}
