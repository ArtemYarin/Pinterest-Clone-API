package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Structures

type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

type IPRateLimiter struct {
	mu                sync.Mutex
	Buckets           map[string]*TokenBucket
	Rate              float64
	Capacity          float64
	TrustProxyHeaders bool
	StopCleanup       chan struct{}
}

// NewIPRateLimiter creates a limiter and starts a background goroutine that
// evicts buckets idle longer than idleTTL, checking every cleanupInterval.
// Call Stop() when you're done with it (e.g. on server shutdown) to release
// the goroutine.
func NewIPRateLimiter(rate, capacity float64, cleanupInterval, idleTTL time.Duration) *IPRateLimiter {
	l := &IPRateLimiter{
		Buckets:     make(map[string]*TokenBucket),
		Rate:        rate,
		Capacity:    capacity,
		StopCleanup: make(chan struct{}),
	}

	go l.cleanupLoop(cleanupInterval, idleTTL)

	return l
}

// Stop terminates the background cleanup goroutine
func (l *IPRateLimiter) Stop() {
	close(l.StopCleanup)
}

// Background task for cleanup
func (l *IPRateLimiter) cleanupLoop(interval, idleTTL time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.evictIdle(idleTTL)
		case <-l.StopCleanup:
			return
		}
	}
}

// evictIdle cleans old buckets whos idle time is bigger than idleTTL
func (l *IPRateLimiter) evictIdle(idleTTL time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	for ip, bucket := range l.Buckets {
		bucket.mu.Lock()
		idle := now.Sub(bucket.lastSeen)
		bucket.mu.Unlock()

		if idle > idleTTL {
			delete(l.Buckets, ip)
		}
	}
}

// getBucket returns the bucket for ip, creating it if needed
func (l *IPRateLimiter) getBucket(ip string) *TokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, exists := l.Buckets[ip]
	if !exists {
		bucket = &TokenBucket{
			tokens:     l.Capacity,
			lastRefill: time.Now(),
		}
		l.Buckets[ip] = bucket
	}
	return bucket
}

// Allow refills the bucket and, if a token is available, consumes it.
// It reports whether the request should proceed, along with a suggested
// Retry-After duration to use when it should not.
func (b *TokenBucket) Allow(rate, capacity float64) (bool, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()

	b.tokens += rate * elapsed
	if b.tokens > capacity {
		b.tokens = capacity
	}
	b.lastRefill = now
	b.lastSeen = now

	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}

	// Not enough tokens: estimate wait time until one becomes available.
	deficit := 1 - b.tokens
	waitSeconds := deficit / rate
	return false, time.Duration(waitSeconds * float64(time.Second))
}

// clientIP extracts the client's IP address from the request.
//
// If TrustProxyHeaders is set on the limiter, it prefers X-Forwarded-For /
// X-Real-IP; otherwise it always uses RemoteAddr.
func clientIP(r *http.Request, trustProxyHeaders bool) string {
	if trustProxyHeaders {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
		if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			return xrip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitingMiddleware enforces the rate limit per client IP.
// If not allowed sets Retry-After header
func (l *IPRateLimiter) RateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r, l.TrustProxyHeaders)
		bucket := l.getBucket(ip)

		allowed, retryAfter := bucket.Allow(l.Rate, l.Capacity)

		// Set headers
		tokensToFull := l.Capacity - bucket.tokens
		timeToFill := tokensToFull / l.Rate
		rateLimitReset := time.Now().Unix() + int64(timeToFill)
		w.Header().Set("RateLimit-Limit", fmt.Sprintf("%f", l.Capacity))
		w.Header().Set("RateLimit-Remaining", fmt.Sprintf("%d", int(bucket.tokens)))
		w.Header().Set("RateLimit-Reset", fmt.Sprintf("%d", rateLimitReset))

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "429 Too Many Requests",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
