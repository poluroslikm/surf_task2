import { useRef, useState } from 'react'
import { authApi } from '../../api/authApi'
import { session } from '../../session/session'
import { ApiRequestError, NetworkError, GENERIC_NETWORK_ERROR, GENERIC_SERVER_ERROR } from '../../core/errors'

export type AuthTab = 'login' | 'register'

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

// SCR-001-auth.md §8 / LOGIC-001. Every text below is quoted verbatim from those documents.
export function useAuthScreen(onAuthenticated: () => void) {
  const [activeTab, setActiveTab] = useState<AuthTab>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [passwordVisible, setPasswordVisible] = useState(false)
  const [emailError, setEmailError] = useState<string | null>(null)
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const passwordInputRef = useRef<HTMLInputElement>(null)

  function selectTab(tab: AuthTab) {
    setActiveTab(tab)
    // §6.1: switching tabs clears the password field (never carry a password across
    // login/register by mistake) but keeps the entered email.
    setPassword('')
    setPasswordError(null)
    setFormError(null)
  }

  function changeEmail(value: string) {
    setEmail(value)
    setEmailError(null)
  }

  function blurEmail() {
    if (email !== '' && !EMAIL_RE.test(email)) {
      setEmailError('Введите корректный email') // E1
    }
  }

  function changePassword(value: string) {
    setPassword(value)
    setPasswordError(null)
  }

  function blurPassword() {
    // §6.3: the length check only applies on the Register tab — Login defers entirely to the
    // server (E3), since an existing account may predate this length policy.
    if (activeTab === 'register' && password !== '' && password.length < 8) {
      setPasswordError('Пароль должен быть не короче 8 символов') // E2
    }
  }

  const canSubmit = email !== '' && password !== '' && !submitting

  async function submit() {
    if (!canSubmit) return
    setSubmitting(true)
    setFormError(null)
    try {
      const response =
        activeTab === 'register' ? await authApi.register({ email, password }) : await authApi.login({ email, password })
      session.setSession(response.token, response.client.email)
      onAuthenticated()
      return
    } catch (err) {
      if (err instanceof NetworkError) {
        setFormError(GENERIC_NETWORK_ERROR) // E5
      } else if (err instanceof ApiRequestError) {
        if (activeTab === 'register' && err.status === 409) {
          // LOGIC-001: this domain case is told apart by HTTP status 409, not by Error.code
          // (the contract reuses "bad_request" for it) — see AC-003/AC-N04.
          setFormError('Этот email уже зарегистрирован. Войдите или используйте другой email.') // E4
        } else if (activeTab === 'login' && err.status === 401) {
          setFormError('Неверный email или пароль') // E3
          setPassword('')
          passwordInputRef.current?.focus()
        } else if (err.status >= 500) {
          setFormError(GENERIC_SERVER_ERROR) // E6
        } else {
          setFormError(err.body.message)
        }
      } else {
        setFormError(GENERIC_SERVER_ERROR)
      }
    } finally {
      setSubmitting(false)
    }
  }

  return {
    activeTab,
    email,
    password,
    passwordVisible,
    emailError,
    passwordError,
    formError,
    submitting,
    canSubmit,
    passwordInputRef,
    selectTab,
    changeEmail,
    blurEmail,
    changePassword,
    blurPassword,
    togglePasswordVisible: () => setPasswordVisible((v) => !v),
    submit,
  }
}
