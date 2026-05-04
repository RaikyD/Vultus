import { useEffect, useState } from 'react'
import { streamsApi } from '../api/streams'
import type { Stream } from '../api/types'
import { StreamCard } from '../components/StreamCard'

export function Home() {
  const [streams, setStreams] = useState<Stream[]>([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    streamsApi.list({ search: search || undefined })
      .then(setStreams)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [search])

  return (
    <div style={{ padding: 24 }}>
      <h1 style={{ marginBottom: 16 }}>Live Streams</h1>

      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search streams or channels..."
        style={{ padding: '8px 12px', borderRadius: 4, border: '1px solid #444', background: '#111', color: '#fff', width: 300, marginBottom: 24 }}
      />

      {loading ? (
        <div>Loading...</div>
      ) : streams.length === 0 ? (
        <div style={{ color: '#888' }}>No streams found.</div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 20 }}>
          {streams.map((s) => <StreamCard key={s.id} stream={s} />)}
        </div>
      )}
    </div>
  )
}
