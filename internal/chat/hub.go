package chat

import (
	"sync"

	"github.com/google/uuid"
)

// Hub manages all chat rooms keyed by stream ID.
type Hub struct {
	mu    sync.RWMutex
	rooms map[uuid.UUID]*room
}

type room struct {
	clients map[*Client]bool
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[uuid.UUID]*room)}
}

// Register adds a client to its stream's room.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	r, ok := h.rooms[c.streamID]
	if !ok {
		r = &room{clients: make(map[*Client]bool)}
		h.rooms[c.streamID] = r
	}
	h.mu.Unlock()

	r.mu.Lock()
	r.clients[c] = true
	r.mu.Unlock()
}

// Unregister removes a client and closes its send channel.
func (h *Hub) Unregister(c *Client) {
	h.mu.RLock()
	r, ok := h.rooms[c.streamID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	r.mu.Lock()
	if _, exists := r.clients[c]; exists {
		delete(r.clients, c)
		close(c.send)
	}
	r.mu.Unlock()
}

// Broadcast sends a message to all clients in a stream room.
func (h *Hub) Broadcast(streamID uuid.UUID, msg []byte) {
	h.mu.RLock()
	r, ok := h.rooms[streamID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		select {
		case c.send <- msg:
		default:
			// Slow client — drop the message rather than blocking
		}
	}
}
