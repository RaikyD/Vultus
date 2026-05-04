import { Link, useNavigate } from 'react-router-dom'
import { authApi } from '../api/auth'
import { useAuthStore } from '../store/auth'

export function Navbar() {
  const { user, setUser } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await authApi.logout()
    setUser(null)
    navigate('/')
  }

  return (
    <nav style={{ padding: '12px 24px', borderBottom: '1px solid #333', display: 'flex', alignItems: 'center', gap: 16 }}>
      <Link to="/" style={{ fontWeight: 700, fontSize: 20, textDecoration: 'none' }}>Sequi</Link>
      <div style={{ flex: 1 }} />
      {user ? (
        <>
          <Link to={`/channel/${user.username}`}>My Channel</Link>
          <Link to="/dashboard">Dashboard</Link>
          <button onClick={handleLogout}>Logout</button>
        </>
      ) : (
        <>
          <Link to="/login">Login</Link>
          <Link to="/register">Register</Link>
        </>
      )}
    </nav>
  )
}
