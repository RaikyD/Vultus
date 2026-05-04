import { useEffect, useState } from 'react'
import { streamKeysApi } from '../api/streams'
import type { StreamKey } from '../api/types'
import { ApiError } from '../api/client'
import { useAuthStore } from '../store/auth'

export function Dashboard() {
  const { user } = useAuthStore()
  const [streamKey, setStreamKey] = useState<StreamKey | null>(null)
  const [loading, setLoading] = useState(true)
  const [regenLoading, setRegenLoading] = useState(false)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    streamKeysApi.get()
      .then(setStreamKey)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [])

  const handleRegen = async () => {
    if (!confirm('Regenerate stream key? OBS will need to be updated.')) return
    setRegenLoading(true)
    try {
      const newKey = await streamKeysApi.regenerate()
      setStreamKey(newKey)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed')
    } finally {
      setRegenLoading(false)
    }
  }

  const handleCopy = async (text: string) => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  if (loading) return <div style={{ padding: 24 }}>Loading...</div>

  const rtmpUrl = `rtmp://localhost:1935/live`

  return (
    <div style={{ maxWidth: 700, margin: '0 auto', padding: 24 }}>
      <h1 style={{ marginBottom: 24 }}>Stream Dashboard</h1>

      {/* OBS Setup */}
      <section style={{ marginBottom: 32 }}>
        <h2 style={{ marginBottom: 16 }}>OBS Setup</h2>

        <div style={{ marginBottom: 16 }}>
          <label style={{ display: 'block', marginBottom: 6, color: '#aaa', fontSize: 13 }}>RTMP Server</label>
          <div style={{ display: 'flex', gap: 8 }}>
            <input
              readOnly
              value={rtmpUrl}
              style={inputStyle}
            />
            <button onClick={() => handleCopy(rtmpUrl)}>Copy</button>
          </div>
        </div>

        <div style={{ marginBottom: 16 }}>
          <label style={{ display: 'block', marginBottom: 6, color: '#aaa', fontSize: 13 }}>Stream Key</label>
          <div style={{ display: 'flex', gap: 8 }}>
            <input
              readOnly
              value={streamKey?.key ?? ''}
              type="text"
              style={inputStyle}
            />
            <button onClick={() => streamKey && handleCopy(streamKey.key)}>
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
        </div>

        <button onClick={handleRegen} disabled={regenLoading} style={{ color: '#e53' }}>
          {regenLoading ? 'Regenerating...' : 'Regenerate Stream Key'}
        </button>

        {error && <div style={{ color: '#e53', marginTop: 8 }}>{error}</div>}
      </section>

      {/* Quick link */}
      {user && (
        <section>
          <h2 style={{ marginBottom: 12 }}>Your Channel</h2>
          <a href={`/channel/${user.username}`}>View public channel page →</a>
        </section>
      )}
    </div>
  )
}

const inputStyle: React.CSSProperties = {
  flex: 1,
  padding: '8px 12px',
  borderRadius: 4,
  border: '1px solid #444',
  background: '#111',
  color: '#fff',
  fontFamily: 'monospace',
  fontSize: 13,
}
