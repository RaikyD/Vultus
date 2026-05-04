package chat

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"

	db "sequi/internal/db/gen"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	sendBufSize    = 64
)

// Client represents a single WebSocket connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	queries  *db.Queries
	limiter  *RateLimiter
	streamID uuid.UUID
	userID   uuid.UUID // zero UUID if unauthenticated
	username string
}

func NewClient(hub *Hub, conn *websocket.Conn, queries *db.Queries, limiter *RateLimiter,
	streamID, userID uuid.UUID, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBufSize),
		queries:  queries,
		limiter:  limiter,
		streamID: streamID,
		userID:   userID,
		username: username,
	}
}

// ReadPump reads messages from the WebSocket connection.
// Must be run in a goroutine. Handles registration cleanup on exit.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var in InboundMessage
		if err := json.Unmarshal(msgBytes, &in); err != nil {
			continue
		}

		if in.Type != TypeMessage {
			continue
		}

		// Only authenticated users can send
		if c.userID == uuid.Nil {
			c.sendError("authentication required to send messages")
			continue
		}

		content := in.Content
		if len(content) == 0 || len(content) > 500 {
			c.sendError("message must be 1–500 characters")
			continue
		}

		if !c.limiter.Allow(c.userID, c.streamID) {
			c.sendError("rate limit exceeded")
			continue
		}

		pgStreamID := pgtype.UUID{Bytes: c.streamID, Valid: true}
		pgUserID := pgtype.UUID{Bytes: c.userID, Valid: true}

		saved, err := c.queries.CreateChatMessage(ctx, db.CreateChatMessageParams{
			StreamID: pgStreamID,
			UserID:   pgUserID,
			Content:  content,
		})
		if err != nil {
			log.Printf("chat: save message: %v", err)
			continue
		}

		out := OutboundMessage{
			Type:      TypeMessage,
			ID:        uuid.UUID(saved.ID.Bytes).String(),
			UserID:    c.userID.String(),
			Username:  c.username,
			Content:   saved.Content,
			CreatedAt: saved.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
		payload, _ := json.Marshal(out)
		c.hub.Broadcast(c.streamID, payload)
	}
}

// WritePump writes messages from the send channel to the WebSocket connection.
// Must be run in a goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send returns the client's outbound send channel.
func (c *Client) Send() chan []byte { return c.send }

func (c *Client) sendError(msg string) {
	payload, _ := json.Marshal(OutboundMessage{Type: TypeError, Message: msg})
	select {
	case c.send <- payload:
	default:
	}
}
