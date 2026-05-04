import { api } from './client'
import type { Category, HLSInfo, Stream, StreamerProfile, StreamKey, Subscription } from './types'

export const streamsApi = {
  list: (params?: { status?: string; search?: string }) => {
    const qs = new URLSearchParams()
    if (params?.status) qs.set('status', params.status)
    if (params?.search) qs.set('search', params.search)
    const suffix = qs.toString() ? `?${qs}` : ''
    return api.get<Stream[]>(`/api/streams${suffix}`)
  },
  getById: (id: string) => api.get<Stream>(`/api/streams/${id}`),
  getHLS: (id: string) => api.get<HLSInfo>(`/api/streams/${id}/hls`),
  categories: () => api.get<Category[]>('/api/categories'),
}

export const streamKeysApi = {
  get: () => api.get<StreamKey>('/api/stream-keys'),
  regenerate: () => api.post<StreamKey>('/api/stream-keys/regenerate'),
}

export const usersApi = {
  getProfile: (username: string) => api.get<StreamerProfile>(`/api/users/${username}`),
  subscribe: (username: string) => api.post<{ subscribed: boolean }>(`/api/users/${username}/subscribe`),
  unsubscribe: (username: string) => api.delete<{ subscribed: boolean }>(`/api/users/${username}/subscribe`),
  mySubscriptions: () => api.get<Subscription[]>('/api/me/subscriptions'),
}
