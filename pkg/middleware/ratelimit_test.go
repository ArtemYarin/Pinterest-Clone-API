package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsBurstThenBlocks(t *testing.T) {
	limiter := &IPRateLimiter{
		Buckets:  make(map[string]*TokenBucket),
		Rate:     5,
		Capacity: 10,
	}

	handler := limiter.RateLimitingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Fire 10 requests instantly from the same "client" — all should succeed.
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		req.RemoteAddr = "1.2.3.4:5555" // same IP every time
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rec.Code)
		}
	}

	// The 11th request should be rejected — burst capacity exhausted.
	req := httptest.NewRequest("GET", "/hello", nil)
	req.RemoteAddr = "1.2.3.4:5555"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("request 11: expected 429, got %d", rec.Code)
	}

	// After waiting, tokens should have refilled and a request succeeds again.
	time.Sleep(300 * time.Millisecond) // ~1.5 tokens at rate=5/sec

	req = httptest.NewRequest("GET", "/hello", nil)
	req.RemoteAddr = "1.2.3.4:5555"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("after refill wait: expected 200, got %d", rec.Code)
	}
}

func TestRateLimiter_DifferentIPsHaveSeparateBuckets(t *testing.T) {
	limiter := &IPRateLimiter{
		Buckets:  make(map[string]*TokenBucket),
		Rate:     5,
		Capacity: 2,
	}
	handler := limiter.RateLimitingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust client A's bucket.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		req.RemoteAddr = "1.1.1.1:1111"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	reqA := httptest.NewRequest("GET", "/hello", nil)
	reqA.RemoteAddr = "1.1.1.1:1111"
	recA := httptest.NewRecorder()
	handler.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusTooManyRequests {
		t.Errorf("client A should be rate limited, got %d", recA.Code)
	}

	// Client B, a different IP, should be unaffected.
	reqB := httptest.NewRequest("GET", "/hello", nil)
	reqB.RemoteAddr = "2.2.2.2:2222"
	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Errorf("client B should not be rate limited, got %d", recB.Code)
	}
}
