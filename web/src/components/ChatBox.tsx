import { FormEvent, useRef, useState } from 'react'
import { useAuthStore } from '../store/auth'
import { useChat } from '../hooks/useChat'

interface Props {
  streamId: string
}

export function ChatBox({ streamId }: Props) {
  const { messages, error, sendMessage } = useChat(streamId)
  const { user } = useAuthStore()
  const [draft, setDraft] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const content = draft.trim()
    if (!content) return
    sendMessage(content)
    setDraft('')
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', border: '1px solid #333', borderRadius: 8 }}>
      <div style={{ padding: '8px 12px', borderBottom: '1px solid #333', fontWeight: 600 }}>Chat</div>

      <div style={{ flex: 1, overflowY: 'auto', padding: '8px 12px', display: 'flex', flexDirection: 'column', gap: 6 }}>
        {messages.map((msg) => (
          <div key={msg.id}>
            <span style={{ fontWeight: 600, marginRight: 8 }}>{msg.username}</span>
            <span>{msg.content}</span>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {error && (
        <div style={{ padding: '4px 12px', color: '#e53', fontSize: 12 }}>{error}</div>
      )}

      {user ? (
        <form onSubmit={handleSubmit} style={{ display: 'flex', gap: 8, padding: '8px 12px', borderTop: '1px solid #333' }}>
          <input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            placeholder="Send a message..."
            maxLength={500}
            style={{ flex: 1, padding: '6px 8px', borderRadius: 4, border: '1px solid #444', background: '#111', color: '#fff' }}
          />
          <button type="submit" disabled={!draft.trim()}>Send</button>
        </form>
      ) : (
        <div style={{ padding: '8px 12px', borderTop: '1px solid #333', fontSize: 13, color: '#888' }}>
          Log in to chat
        </div>
      )}
    </div>
  )
}
