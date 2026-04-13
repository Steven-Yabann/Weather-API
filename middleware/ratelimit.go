package middleware

import (
	"net/http"
	"sync"
	"time"
)

type limiter struct {
	tokens		float64
	maxTokens	float64
	refillPS	float64	// tokens per second
	lastTime	time.Time
	mu			sync.Mutex
}

func (l *limiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	// Refill
	l.tokens = min(l.maxTokens, l.tokens + elapsed * l.refillPS)

	if l.tokens < 1 { return false }

	l.tokens--

	return true
}

type RateLimiter struct {
	mu			sync.Mutex
	limiters	map[string]*limiter
	maxTokens 	float64
	refillPS	float64
}

func NewRateLimiter() *RateLimiter {
	r1 := &RateLimiter {
		limiters:	make(map[string]*limiter),
		maxTokens: 	5,
		refillPS: 	1,
	}

	// Periodically clean up stale IPs
	go r1.cleanup()

	return r1
}

func (r1 *RateLimiter) getLimiter (ip string) *limiter {
	r1.mu.Lock()
	defer r1.mu.Unlock()

	l, exists := r1.limiters[ip]
	
	if !exists {
		l = &limiter {
			tokens:		r1.maxTokens,
			maxTokens: 	r1.maxTokens,
			refillPS: 	r1.refillPS,
			lastTime: 	time.Now(),
		}
		r1.limiters[ip] = l
	}

	return l
}

func (r1 *RateLimiter) Middleware (next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		if !r1.getLimiter(ip).allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (r1 *RateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		r1.mu.Lock()
		r1.limiters = make(map[string]*limiter)
		r1.mu.Unlock()
	}
}

func min (a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}