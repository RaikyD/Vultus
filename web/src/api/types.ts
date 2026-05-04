export interface User {
  id: string
  email: string
  username: string
  avatar_url: string | null
  bio: string | null
  created_at: string
}

export interface Stream {
  id: string
  user_id: string
  username: string
  user_avatar_url: string | null
  category_id: string | null
  title: string
  status: 'live' | 'offline'
  viewer_count: number
  thumbnail_url: string | null
  started_at: string | null
}

export interface Category {
  id: string
  name: string
}

export interface StreamKey {
  id: string
  key: string
  created_at: string
}

export interface StreamerProfile {
  id: string
  username: string
  avatar_url: string | null
  bio: string | null
  subscriber_count: number
  is_subscribed: boolean
  stream: {
    id: string
    title: string
    status: 'live' | 'offline'
    viewer_count: number
  }
}

export interface Subscription {
  id: string
  username: string
  avatar_url: string | null
  subscribed_at: string
}

export interface HLSInfo {
  url: string
}
