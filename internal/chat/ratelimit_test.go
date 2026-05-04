package chat

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRateLimiter_AllowsUpToMax(t *testing.T) {
	rl := NewRateLimiter()
	userID := uuid.New()
	streamID := uuid.New()

	for i := range rateMaxMsgs {
		if !rl.Allow(userID, streamID) {
			t.Fatalf("expected Allow=true on message %d", i+1)
		}
	}
}

func TestRateLimiter_BlocksAfterMax(t *testing.T) {
	rl := NewRateLimiter()
	userID := uuid.New()
	streamID := uuid.New()

	for range rateMaxMsgs {
		rl.Allow(userID, streamID)
	}

	if rl.Allow(userID, streamID) {
		t.Error("expected Allow=false after exceeding rate limit")
	}
}

func TestRateLimiter_IndependentPerUser(t *testing.T) {
	rl := NewRateLimiter()
	streamID := uuid.New()
	u1, u2 := uuid.New(), uuid.New()

	// Exhaust u1's bucket
	for range rateMaxMsgs {
		rl.Allow(u1, streamID)
	}

	// u2 should still be allowed
	if !rl.Allow(u2, streamID) {
		t.Error("u2 should not be affected by u1's rate limit")
	}
}

func TestRateLimiter_IndependentPerStream(t *testing.T) {
	rl := NewRateLimiter()
	userID := uuid.New()
	s1, s2 := uuid.New(), uuid.New()

	// Exhaust bucket for stream s1
	for range rateMaxMsgs {
		rl.Allow(userID, s1)
	}

	// Same user in stream s2 should be allowed
	if !rl.Allow(userID, s2) {
		t.Error("rate limit should be per (user, stream), not just per user")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	// Override window for test speed — we do this by checking the exported const
	// is > 0 and just verifying the reset logic internally.
	rl := &RateLimiter{buckets: make(map[string]*bucket)}
	userID := uuid.New()
	streamID := uuid.New()
	key := userID.String() + ":" + streamID.String()

	// Manually set an expired bucket
	rl.mu.Lock()
	rl.buckets[key] = &bucket{count: rateMaxMsgs, windowEnd: time.Now().Add(-time.Second)}
	rl.mu.Unlock()

	// After window expiry, Allow should reset and return true
	if !rl.Allow(userID, streamID) {
		t.Error("expected Allow=true after window expiry")
	}

	rl.mu.Lock()
	b := rl.buckets[key]
	rl.mu.Unlock()

	if b.count != 1 {
		t.Errorf("expected count=1 after reset, got %d", b.count)
	}
}
