import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { usersApi } from '../api/streams'
import type { StreamerProfile } from '../api/types'
import { useAuthStore } from '../store/auth'
import { ApiError } from '../api/client'

export function Channel() {
  const { username } = useParams<{ username: string }>()
  const { user } = useAuthStore()
  const [profile, setProfile] = useState<StreamerProfile | null>(null)
  const [subLoading, setSubLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!username) return
    usersApi.getProfile(username).then(setProfile).catch(console.error)
  }, [username])

  const toggleSubscribe = async () => {
    if (!profile) return
    setSubLoading(true)
    try {
      if (profile.is_subscribed) {
        await usersApi.unsubscribe(profile.username)
        setProfile({ ...profile, is_subscribed: false, subscriber_count: profile.subscriber_count - 1 })
      } else {
        await usersApi.subscribe(profile.username)
        setProfile({ ...profile, is_subscribed: true, subscriber_count: profile.subscriber_count + 1 })
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed')
    } finally {
      setSubLoading(false)
    }
  }

  if (!profile) return <div style={{ padding: 24 }}>Loading...</div>

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: 24 }}>
      {/* Profile header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 20, marginBottom: 32 }}>
        <div style={{ width: 80, height: 80, borderRadius: '50%', background: '#333', overflow: 'hidden', flexShrink: 0 }}>
          {profile.avatar_url && <img src={profile.avatar_url} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />}
        </div>
        <div style={{ flex: 1 }}>
          <h2 style={{ margin: 0 }}>{profile.username}</h2>
          <div style={{ color: '#888', margin: '4px 0' }}>{profile.subscriber_count} subscribers</div>
          {profile.bio && <p style={{ margin: 0, color: '#ccc' }}>{profile.bio}</p>}
        </div>
        {user && user.username !== profile.username && (
          <button onClick={toggleSubscribe} disabled={subLoading}>
            {profile.is_subscribed ? 'Unsubscribe' : 'Subscribe'}
          </button>
        )}
      </div>

      {error && <div style={{ color: '#e53', marginBottom: 16 }}>{error}</div>}

      {/* Stream status */}
      <div style={{ border: '1px solid #333', borderRadius: 8, padding: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
          <span style={{
            padding: '2px 10px',
            borderRadius: 4,
            background: profile.stream.status === 'live' ? '#e53' : '#444',
            fontWeight: 700,
            fontSize: 13,
          }}>
            {profile.stream.status === 'live' ? 'LIVE' : 'OFFLINE'}
          </span>
          <span style={{ fontWeight: 600 }}>{profile.stream.title || 'No title set'}</span>
        </div>
        {profile.stream.status === 'live' && (
          <div style={{ color: '#888', fontSize: 13, marginBottom: 12 }}>
            {profile.stream.viewer_count} viewers
          </div>
        )}
        {profile.stream.status === 'live' && (
          <Link to={`/watch/${profile.stream.id}`}>
            <button>Watch Stream</button>
          </Link>
        )}
      </div>
    </div>
  )
}
