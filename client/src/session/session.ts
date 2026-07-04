// LOGIC-001 → «Хранение и использование токена»: localStorage (survives tab/browser close),
// not sessionStorage — deliberate per the spec, not an oversight.
const TOKEN_KEY = 'chef-stol.token'

const listeners = new Set<() => void>()

function notify() {
  listeners.forEach((l) => l())
}

export const session = {
  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY)
  },
  setToken(token: string) {
    localStorage.setItem(TOKEN_KEY, token)
    notify()
  },
  // Explicit logout (BS-004) and centralized 401-on-any-request-but-login handling
  // (LOGIC-001 AC-007) both funnel through here.
  clear() {
    localStorage.removeItem(TOKEN_KEY)
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
