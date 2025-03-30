package tinyurl

import (
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Allow(ip string) bool
	Enabled() bool
}

type TokenBucketRateLimiter struct {
	limit   rate.Limit
	burst   int
	enabled bool

	mu      sync.RWMutex
	clients map[string]*rate.Limiter
}

func NewTokenBucketRateLimiter(maxRequestRate float64, requestBurstLimit int, enabled bool) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		limit:   rate.Limit(maxRequestRate),
		burst:   requestBurstLimit,
		enabled: enabled,
		clients: make(map[string]*rate.Limiter),
	}
}

func (l *TokenBucketRateLimiter) Allow(ip string) bool {
	l.mu.RLock()
	rl, ok := l.clients[ip]
	l.mu.RUnlock()

	if !ok {
		rl = rate.NewLimiter(l.limit, l.burst)

		l.mu.Lock()
		l.clients[ip] = rl
		l.mu.Unlock()
	}

	return rl.Allow()
}

func (l *TokenBucketRateLimiter) Enabled() bool {
	return l.enabled
}
