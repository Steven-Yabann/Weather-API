package middleware

import (
	"net/http"
	"sync"
	"time"
)

type limiter struct {
	tokens		float64		// current no of permissions available to make requests
	maxTokens	float64		// max no of permissions
	refillPS	float64		// tokens per second
	lastTime	time.Time	// last time tokens were refilled. Used to calculate tokens to add
	lastSeen	time.Time	// Tracks last request time
	mu			sync.Mutex	// mutex to prevent race conditions from multiple web requests
}

/* 
	process
	- lock the mutex
	- calculate how many tokens were earned since lastTime
	- check token >= 1
	- Decrement token and Unlock() mutex
*/

func (l *limiter) allow() bool {
	// lock the mutex
	l.mu.Lock()
	// defer will unlock the mutex when we exit the function call
	defer l.mu.Unlock()

	// Calculate tokens
	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds() // used to calculate how many tokens were earned
	l.lastTime = now	// update lastTime

	// Refill
	// use min to get the smaller value between maxTokens and tokens earned
	l.tokens = min(l.maxTokens, l.tokens + elapsed * l.refillPS)

	// Check if request is allowed
	if l.tokens < 1 { return false }

	l.tokens--

	return true
}

//	Manager for multiple individual limiters
type RateLimiter struct {
	mu			sync.Mutex				// protects limiters map from crashing
	limiters	map[string]*limiter		// storage. key is identifier and value is pointer to limiter pointer
	maxTokens 	float64					// template values. Used to create new limiters
	refillPS	float64					// template values. Used to create new limiters
}

//	RateLimiter constructor
func NewRateLimiter() *RateLimiter {
	// create a new RateLimiter
	r1 := &RateLimiter {
		limiters:	make(map[string]*limiter),	// create the map
		maxTokens: 	5,							// template values to maxTokens
		refillPS: 	1,							// template values to refillPS
	}

	// Periodically clean up stale IPs
	go r1.cleanup()

	return r1
}

//	getLimiter returns the limiter for the given IP
func (r1 *RateLimiter) getLimiter (ip string) *limiter {
	// lock the RateLimiter object
	r1.mu.Lock()
	defer r1.mu.Unlock()

	// get the limiter
	l, exists := r1.limiters[ip]
	
	// if it doesnt exist, create a new limiter for that IP
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

// 	before the request reaches the WeatherAPI, it passes through this middleware
//	parameter is the next handler
func (r1 *RateLimiter) Middleware (next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr	// retreives the IP of the request

		//	if the .allow() returns false, return error for too many requests
		if !r1.getLimiter(ip).allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		//	otherwise, pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

func (r1 *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r1.mu.Lock()

		defer r1.mu.Unlock()

		for ip, l := range r1.limiters {
			// if a user hasnt made a request in 3 minutes, remove them
			if time.Since(l.lastSeen) > 3 * time.Minute {
				delete(r1.limiters, ip)
			}
		}
	}
}

func min (a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}