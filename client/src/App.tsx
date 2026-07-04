import { useEffect, useState } from 'react'
import { session } from './session/session'
import { AuthScreen } from './screens/auth/AuthScreen'
import { SlotsScreen } from './screens/slots/SlotsScreen'

// Root: simple Auth <-> Slots switch on session presence — no router yet (only 2 screens exist
// so far; full tab-bar navigation is a separate, not-yet-started item, see
// 6-development/FE_IMPLEMENTATION_PLAN.md FE-05).
function App() {
  const [authenticated, setAuthenticated] = useState(session.isAuthenticated())

  useEffect(() => session.subscribe(() => setAuthenticated(session.isAuthenticated())), [])

  if (!authenticated) {
    return <AuthScreen onAuthenticated={() => setAuthenticated(true)} />
  }

  return (
    <SlotsScreen
      onSlotSelected={(slotId) => {
        // SCR-003 doesn't exist yet.
        console.log('slot selected', slotId)
      }}
    />
  )
}

export default App
