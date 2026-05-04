import { Link } from 'react-router-dom'
import type { Stream } from '../api/types'

interface Props {
  stream: Stream
}

export function StreamCard({ stream }: Props) {
  return (
    <Link to={`/watch/${stream.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
      <div style={{ border: '1px solid #333', borderRadius: 8, overflow: 'hidden' }}>
        <div style={{ background: '#111', aspectRatio: '16/9', display: 'flex', alignItems: 'center', justifyContent: 'center', position: 'relative' }}>
          {stream.thumbnail_url ? (
            <img src={stream.thumbnail_url} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
          ) : (
            <span style={{ color: '#555' }}>No preview</span>
          )}
          {stream.status === 'live' && (
            <span style={{ position: 'absolute', top: 8, left: 8, background: '#e53', color: '#fff', padding: '2px 8px', borderRadius: 4, fontSize: 12, fontWeight: 700 }}>
              LIVE
            </span>
          )}
        </div>
        <div style={{ padding: 12 }}>
          <div style={{ fontWeight: 600, marginBottom: 4, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {stream.title || 'Untitled stream'}
          </div>
          <div style={{ fontSize: 13, color: '#aaa' }}>
            <Link to={`/channel/${stream.username}`} onClick={(e) => e.stopPropagation()}>
              {stream.username}
            </Link>
          </div>
          {stream.status === 'live' && (
            <div style={{ fontSize: 12, color: '#888', marginTop: 4 }}>
              {stream.viewer_count} viewers
            </div>
          )}
        </div>
      </div>
    </Link>
  )
}
