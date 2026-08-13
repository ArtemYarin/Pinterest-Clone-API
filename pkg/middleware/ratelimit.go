package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Structures

type TokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	Buckets  map[string]*TokenBucket
	Rate     float64
	Capacity float64
}

// GetBucket returns TokenBucket with sender's ID, if
// it doesn't exist, it will be created.
func (ipLimiter *IPRateLimiter) GetBucket(ip string) *TokenBucket {
	ipLimiter.mu.Lock()
	defer ipLimiter.mu.Unlock()

	bucket, exists := ipLimiter.Buckets[ip]
	if !exists {
		bucket = &TokenBucket{
			tokens:     ipLimiter.Capacity,
			lastRefill: time.Now(),
		}
		ipLimiter.Buckets[ip] = bucket
	}
	return bucket
}

// Refilles tokens in TokenBucket.
// rate is amount of tokens refill per second.
// capacity is burst size
func (tBucket *TokenBucket) Refill(rate, capacity float64) {
	now := time.Now()
	elapsed := now.Sub(tBucket.lastRefill).Seconds()
	// Refill
	tBucket.tokens += rate * elapsed

	// Capacity check
	if tBucket.tokens > capacity {
		tBucket.tokens = capacity
	}

	// Set lastRefill
	tBucket.lastRefill = now
}

func (tBucket *TokenBucket) Allow(rate, capacity float64) bool {
	tBucket.Refill(rate, capacity)

	if tBucket.tokens >= 1 {
		tBucket.tokens--
		return true
	}
	return false
}

// Gets user IP
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Middleware
func (ipLimiter *IPRateLimiter) RateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		bucket := ipLimiter.GetBucket(ip)

		if !bucket.Allow(ipLimiter.Rate, ipLimiter.Capacity) {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
