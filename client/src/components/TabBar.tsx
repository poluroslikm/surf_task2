import { Link, useLocation } from 'react-router-dom'

// foundations §4.2: two root tabs, visible on the two root screens (SCR-002/SCR-005), hidden
// on nested screens (slot card, booking form, booking details, rating) — those screens simply
// don't render this component, matching the existing SlotsScreen structure.
export function TabBar() {
  const location = useLocation()
  const onBookings = location.pathname.startsWith('/bookings')

  return (
    <nav className="tab-bar">
      <Link to="/" className={`tab-bar__item ${!onBookings ? 'tab-bar__item--active' : ''}`}>
        🍳 Классы
      </Link>
      <Link to="/bookings" className={`tab-bar__item ${onBookings ? 'tab-bar__item--active' : ''}`}>
        🗓 Мои записи
      </Link>
    </nav>
  )
}
