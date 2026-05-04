import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { streamsApi, usersApi } from '../api/streams'
import type { HLSInfo, Stream, StreamerProfile } from '../api/types'
import { ChatBox } from '../components/ChatBox'
import { useAuthStore } from '../store/auth'
import { useHLS } from '../hooks/useHLS'
import { ApiError } from '../api/client'

export function Watch() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuthStore()
  const [stream, setStream] = useState<Stream | null>(null)
  const [profile, setProfile] = useState<StreamerProfile | null>(null)
  const [hlsInfo, setHlsInfo] = useState<HLSInfo | null>(null)
  const [subLoading, setSubLoading] = useState(false)
  const [error, setError] = useState('')

  const videoRef = useHLS(stream?.status === 'live' ? (hlsInfo?.url ?? null) : null)

  useEffect(() => {
    if (!id) return
    streamsApi.getById(id).then((s) => {
      setStream(s)
      return Promise.all([
        s.status === 'live' ? streamsApi.getHLS(id) : Promise.resolve(null),
        usersApi.getProfile(s.username),
      ])
    }).then(([hls, prof]) => {
      if (hls) setHlsInfo(hls)
      setProfile(prof)
    }).catch(console.error)
  }, [id])

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

  if (!stream) return <div style={{ padding: 24 }}>Loading...</div>

  return (
    <div style={{ padding: 24, display: 'grid', gridTemplateColumns: '1fr 320px', gap: 24, maxWidth: 1200, margin: '0 auto' }}>
      <div>
        {/* Video player */}
        <div style={{ background: '#000', aspectRatio: '16/9', borderRadius: 8, overflow: 'hidden', marginBottom: 16, position: 'relative' }}>
          {stream.status === 'live' ? (
            <>
              <video ref={videoRef} autoPlay muted playsInline style={{ width: '100%', height: '100%', display: 'block' }} />
              {/* Live badge — no seekbar shown */}
              <div style={{ position: 'absolute', bottom: 12, left: 12, background: '#e53', color: '#fff', fontSize: 11, fontWeight: 700, padding: '2px 7px', borderRadius: 3, letterSpacing: 1, textTransform: 'uppercase', pointerEvents: 'none' }}>
                Live
              </div>
            </>
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: '#555' }}>
              Stream is offline
            </div>
          )}
        </div>

        {/* Stream info */}
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: 16 }}>
          <div style={{ flex: 1 }}>
            <h2 style={{ margin: 0 }}>{stream.title || 'Untitled stream'}</h2>
            <div style={{ color: '#888', marginTop: 4 }}>
              {stream.status === 'live' ? `${stream.viewer_count} viewers` : 'Offline'}
            </div>
          </div>

          {profile && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div>
                <div style={{ fontWeight: 600 }}>{profile.username}</div>
                <div style={{ fontSize: 12, color: '#888' }}>{profile.subscriber_count} subscribers</div>
              </div>
              {user && user.username !== profile.username && (
                <button onClick={toggleSubscribe} disabled={subLoading}>
                  {profile.is_subscribed ? 'Unsubscribe' : 'Subscribe'}
                </button>
              )}
            </div>
          )}
        </div>
        {error && <div style={{ color: '#e53', marginTop: 8 }}>{error}</div>}
      </div>

      {/* Chat */}
      <div style={{ height: 'calc(100vh - 120px)', minHeight: 400 }}>
        <ChatBox streamId={stream.id} />
      </div>
    </div>
  )
}
