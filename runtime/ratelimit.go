package runtime

import (
	"fmt"
	"sync"
	"time"
)

type tokenBucket struct {
	mu     sync.Mutex
	tokens float64
	rate   float64
	burst  float64
	last   time.Time
}

func newTokenBucket(rate, burst float64) *tokenBucket {
	return &tokenBucket{tokens: burst, rate: rate, burst: burst, last: time.Now()}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	b.tokens += now.Sub(b.last).Seconds() * b.rate
	b.last = now
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

type rateLimiter struct {
	mu       sync.Mutex
	perSess  map[string]*tokenBucket
	rate     float64
	burst    float64
}

func newRateLimiter(rate, burst float64) *rateLimiter {
	return &rateLimiter{perSess: make(map[string]*tokenBucket), rate: rate, burst: burst}
}

func (r *rateLimiter) allow(sessID string) error {
	r.mu.Lock()
	b, ok := r.perSess[sessID]
	if !ok {
		b = newTokenBucket(r.rate, r.burst)
		r.perSess[sessID] = b
	}
	r.mu.Unlock()
	if !b.allow() {
		return fmt.Errorf("portico: action rate limit exceeded")
	}
	return nil
}
