// Direct reflection of 5-mobile-app-spec/09_Логики/LOGIC-005_Паттерн-состояний-экрана.md —
// not a generic "just in case" abstraction. Loading is a full-screen skeleton shown only when
// there's nothing to display yet; Refreshing keeps prior content visible during pull-to-refresh
// and never falls back to Error (LOGIC-005 AC-007).
export type ScreenState<T> =
  | { kind: 'loading' }
  | { kind: 'content'; value: T; refreshing: boolean }
  | { kind: 'empty'; reason: string }
  | { kind: 'error'; message: string }

export const ScreenState = {
  loading: <T>(): ScreenState<T> => ({ kind: 'loading' }),
  content: <T>(value: T, refreshing = false): ScreenState<T> => ({ kind: 'content', value, refreshing }),
  empty: <T>(reason: string): ScreenState<T> => ({ kind: 'empty', reason }),
  error: <T>(message: string): ScreenState<T> => ({ kind: 'error', message }),
}
