package limiter

import (
    "sync"
    "time"
)

type RateLimiter struct {
    mu      sync.Mutex
    rate    int
    window  time.Duration
    tokens  map[string][]time.Time
}

func New(rate int, windowSeconds int) *RateLimiter {
    return &RateLimiter{
        rate:    rate,
        window:  time.Duration(windowSeconds) * time.Second,
        tokens:  make(map[string][]time.Time),
    }
}

func (l *RateLimiter) Allow(key string) bool {
    l.mu.Lock()
    defer l.mu.Unlock()

    now := time.Now()
    cutoff := now.Add(-l.window)

    // Clean old tokens
    times := l.tokens[key]
    valid := times[:0]
    for _, t := range times {
        if t.After(cutoff) {
            valid = append(valid, t)
        }
    }
    l.tokens[key] = valid

    // Check rate limit
    if len(valid) >= l.rate {
        return false
    }

    // Add new token
    l.tokens[key] = append(l.tokens[key], now)
    return true
}