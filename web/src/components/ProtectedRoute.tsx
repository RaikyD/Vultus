import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, initialized } = useAuthStore()

  if (!initialized) return <div>Loading...</div>
  if (!user) return <Navigate to="/login" replace />

  return <>{children}</>
}
