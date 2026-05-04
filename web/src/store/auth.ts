import { create } from 'zustand'
import type { User } from '../api/types'

interface AuthState {
  user: User | null
  initialized: boolean
  setUser: (user: User | null) => void
  setInitialized: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  initialized: false,
  setUser: (user) => set({ user }),
  setInitialized: () => set({ initialized: true }),
}))
