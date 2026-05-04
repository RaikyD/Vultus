package chat

// Message types used over WebSocket.
const (
	TypeMessage = "chat.message"
	TypeHistory = "chat.history"
	TypeError   = "chat.error"
)

// InboundMessage is what the client sends.
type InboundMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// OutboundMessage is what the server sends to all clients.
type OutboundMessage struct {
	Type      string `json:"type"`
	ID        string `json:"id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Username  string `json:"username,omitempty"`
	Content   string `json:"content,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Message   string `json:"message,omitempty"` // used for chat.error
}

// HistoryMessage is the payload for chat.history.
type HistoryMessage struct {
	Type     string            `json:"type"`
	Messages []OutboundMessage `json:"messages"`
}
