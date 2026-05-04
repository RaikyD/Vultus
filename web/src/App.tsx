import { useEffect } from 'react'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { authApi } from './api/auth'
import { Navbar } from './components/Navbar'
import { ProtectedRoute } from './components/ProtectedRoute'
import { useAuthStore } from './store/auth'
import { Home } from './pages/Home'
import { Login } from './pages/Login'
import { Register } from './pages/Register'
import { Watch } from './pages/Watch'
import { Channel } from './pages/Channel'
import { Dashboard } from './pages/Dashboard'

export default function App() {
  const { setUser, setInitialized } = useAuthStore()

  // Restore session on mount
  useEffect(() => {
    authApi.me()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(setInitialized)
  }, [setUser, setInitialized])

  return (
    <BrowserRouter>
      <div style={{ minHeight: '100vh', background: '#0e0e0e', color: '#fff', fontFamily: 'system-ui, sans-serif' }}>
        <Navbar />
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/watch/:id" element={<Watch />} />
          <Route path="/channel/:username" element={<Channel />} />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          />
        </Routes>
      </div>
    </BrowserRouter>
  )
}
