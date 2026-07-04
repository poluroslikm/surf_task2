import { useAuthScreen } from './useAuthScreen'
import { WaveDivider } from '../../components/Illustrations'

// SCR-001. Structure follows the ASCII wireframe in 3-design-brief/SCR-001-auth.md §5: title ->
// tabs -> email -> password -> error area -> fixed bottom CTA.
export function AuthScreen({ onAuthenticated }: { onAuthenticated: () => void }) {
  const s = useAuthScreen(onAuthenticated)

  return (
    <div className="screen screen--auth">
      <header className="auth-header">
        <div className="auth-header__emoji" aria-hidden="true">
          🍳🥘🌿
        </div>
        <h1>Шеф-стол</h1>
        <WaveDivider color="var(--bg)" />
      </header>

      <div className="screen-scroll">
        <div className="tabs" role="tablist">
          <button
            role="tab"
            aria-selected={s.activeTab === 'login'}
            className={`tab ${s.activeTab === 'login' ? 'tab--active' : ''}`}
            onClick={() => s.selectTab('login')}
          >
            Вход
          </button>
          <button
            role="tab"
            aria-selected={s.activeTab === 'register'}
            className={`tab ${s.activeTab === 'register' ? 'tab--active' : ''}`}
            onClick={() => s.selectTab('register')}
          >
            Регистрация
          </button>
        </div>

        <label className="field">
          <span>Email</span>
          <input
            type="email"
            autoComplete="username"
            value={s.email}
            disabled={s.submitting}
            onChange={(e) => s.changeEmail(e.target.value)}
            onBlur={s.blurEmail}
          />
          {s.emailError && <span className="field-error">{s.emailError}</span>}
        </label>

        <label className="field">
          <span>Пароль</span>
          <div className="password-field">
            <input
              ref={s.passwordInputRef}
              type={s.passwordVisible ? 'text' : 'password'}
              autoComplete={s.activeTab === 'login' ? 'current-password' : 'new-password'}
              value={s.password}
              disabled={s.submitting}
              onChange={(e) => s.changePassword(e.target.value)}
              onBlur={s.blurPassword}
            />
            <button
              type="button"
              className="password-toggle"
              aria-label={s.passwordVisible ? 'Скрыть пароль' : 'Показать пароль'}
              onClick={s.togglePasswordVisible}
            >
              {s.passwordVisible ? 'Скрыть' : 'Показать'}
            </button>
          </div>
          {s.passwordError && <span className="field-error">{s.passwordError}</span>}
        </label>

        {s.formError && <div className="form-error">{s.formError}</div>}
      </div>

      <div className="bottom-cta">
        <button disabled={!s.canSubmit} onClick={() => void s.submit()}>
          {s.submitting ? '…' : s.activeTab === 'register' ? 'Зарегистрироваться' : 'Войти'}
        </button>
      </div>
    </div>
  )
}
