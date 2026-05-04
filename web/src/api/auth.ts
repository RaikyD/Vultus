import { api } from './client'
import type { User } from './types'

export const authApi = {
  me: () => api.get<User>('/api/me'),
  register: (email: string, username: string, password: string) =>
    api.post<User>('/api/auth/register', { email, username, password }),
  login: (email: string, password: string) =>
    api.post<User>('/api/auth/login', { email, password }),
  logout: () => api.post<void>('/api/auth/logout'),
}
