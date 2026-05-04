import { useCallback, useEffect, useRef, useState } from 'react'

export interface ChatMessage {
  id: string
  user_id: string
  username: string
  content: string
  created_at: string
}

interface WsPayload {
  type: 'chat.message' | 'chat.history' | 'chat.error'
  id?: string
  user_id?: string
  username?: string
  content?: string
  created_at?: string
  messages?: ChatMessage[]
  message?: string
}

export function useChat(streamId: string | null) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [error, setError] = useState<string | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    let cancelled = false

    const connect = () => {
      if (cancelled || !streamId) return
      const proto = location.protocol === 'https:' ? 'wss' : 'ws'
      const ws = new WebSocket(`${proto}://${location.host}/ws/streams/${streamId}/chat`)
      wsRef.current = ws

      ws.onmessage = (e: MessageEvent) => {
        const payload: WsPayload = JSON.parse(e.data as string)
        if (payload.type === 'chat.history' && payload.messages) {
          setMessages(payload.messages)
        } else if (payload.type === 'chat.message') {
          const msg: ChatMessage = {
            id: payload.id!,
            user_id: payload.user_id!,
            username: payload.username!,
            content: payload.content!,
            created_at: payload.created_at!,
          }
          setMessages((prev) => [...prev, msg])
        } else if (payload.type === 'chat.error') {
          setError(payload.message ?? 'error')
          setTimeout(() => setError(null), 3000)
        }
      }

      ws.onclose = () => {
        if (!cancelled) {
          reconnectRef.current = setTimeout(connect, 3000)
        }
      }

      ws.onerror = () => ws.close()
    }

    connect()

    return () => {
      cancelled = true
      wsRef.current?.close()
      if (reconnectRef.current) clearTimeout(reconnectRef.current)
    }
  }, [streamId])

  const sendMessage = useCallback((content: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'chat.message', content }))
    }
  }, [])

  return { messages, error, sendMessage }
}
