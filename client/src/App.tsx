import { useEffect, useState } from 'react'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { session } from './session/session'
import { AuthScreen } from './screens/auth/AuthScreen'
import { SlotsScreen } from './screens/slots/SlotsScreen'
import { SlotCardScreen } from './screens/slotCard/SlotCardScreen'
import { BookingScreen } from './screens/booking/BookingScreen'
import { MyBookingsScreen } from './screens/myBookings/MyBookingsScreen'
import { BookingDetailsScreen } from './screens/bookingDetails/BookingDetailsScreen'
import { RatingScreen } from './screens/rating/RatingScreen'

// Root: Auth <-> authenticated router switch on session presence (FE-05). Route contract
// (see 6-development/FE_IMPLEMENTATION_PLAN.md for the full map):
//   /                          -> SlotsScreen (SCR-002)
//   /slots/:slotId             -> SlotCardScreen (SCR-003)
//   /slots/:slotId/book        -> BookingScreen (SCR-004, BS-002 as an overlay on success)
//   /bookings                  -> MyBookingsScreen (SCR-005)
//   /bookings/:bookingId       -> BookingDetailsScreen (SCR-006, BS-003 as an overlay)
//   /bookings/:bookingId/rate  -> RatingScreen (SCR-007)
function App() {
  const [authenticated, setAuthenticated] = useState(session.isAuthenticated())

  useEffect(() => session.subscribe(() => setAuthenticated(session.isAuthenticated())), [])

  if (!authenticated) {
    return <AuthScreen onAuthenticated={() => setAuthenticated(true)} />
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<SlotsScreen />} />
        <Route path="/slots/:slotId" element={<SlotCardScreen />} />
        <Route path="/slots/:slotId/book" element={<BookingScreen />} />
        <Route path="/bookings" element={<MyBookingsScreen />} />
        <Route path="/bookings/:bookingId" element={<BookingDetailsScreen />} />
        <Route path="/bookings/:bookingId/rate" element={<RatingScreen />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
