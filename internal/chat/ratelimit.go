package chat

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	rateWindow   = 10 * time.Second
	rateMaxMsgs  = 5
)

// RateLimiter tracks message counts per user per stream.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count     int
	windowEnd time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{buckets: make(map[string]*bucket)}
}

// Allow returns true if the user is within the rate limit.
func (rl *RateLimiter) Allow(userID, streamID uuid.UUID) bool {
	key := userID.String() + ":" + streamID.String()
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok || now.After(b.windowEnd) {
		rl.buckets[key] = &bucket{count: 1, windowEnd: now.Add(rateWindow)}
		return true
	}
	if b.count >= rateMaxMsgs {
		return false
	}
	b.count++
	return true
}
