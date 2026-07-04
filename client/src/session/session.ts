// LOGIC-001 → «Хранение и использование токена»: localStorage (survives tab/browser close),
// not sessionStorage — deliberate per the spec, not an oversight.
const TOKEN_KEY = 'chef-stol.token'
// BS-004 only ever displays the email read-only ("под кем я вошёл") — not used for any request,
// so storing it alongside the token (rather than re-fetching from a nonexistent profile
// endpoint — there is none, see api/README.md) is the simplest faithful implementation.
const EMAIL_KEY = 'chef-stol.email'

const listeners = new Set<() => void>()

function notify() {
  listeners.forEach((l) => l())
}

export const session = {
  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY)
  },
  getEmail(): string | null {
    return localStorage.getItem(EMAIL_KEY)
  },
  setSession(token: string, email: string) {
    localStorage.setItem(TOKEN_KEY, token)
    localStorage.setItem(EMAIL_KEY, email)
    notify()
  },
  // Explicit logout (BS-004) and centralized 401-on-any-request-but-login handling
  // (LOGIC-001 AC-007) both funnel through here.
  clear() {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(EMAIL_KEY)
    notify()
  },
  isAuthenticated(): boolean {
    return this.getToken() !== null
  },
  subscribe(listener: () => void): () => void {
    listeners.add(listener)
    return () => listeners.delete(listener)
  },
}
