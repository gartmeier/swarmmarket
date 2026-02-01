package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	// Create limiter with 10 RPS and burst of 10
	limiter := NewRateLimiter(10, 10)

	key := "test-key"

	// First 10 requests should succeed (burst)
	for i := 0; i < 10; i++ {
		if !limiter.Allow(key) {
			t.Errorf("request %d should be allowed", i)
		}
	}

	// 11th request should be denied
	if limiter.Allow(key) {
		t.Error("11th request should be denied")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	// Create limiter with high RPS for quick refill
	limiter := NewRateLimiter(100, 1)

	key := "test-key"

	// Use up the token
	if !limiter.Allow(key) {
		t.Error("first request should be allowed")
	}

	// Should be denied immediately
	if limiter.Allow(key) {
		t.Error("second immediate request should be denied")
	}

	// Wait for refill (10ms = 1 token at 100 RPS)
	time.Sleep(15 * time.Millisecond)

	// Should be allowed after refill
	if !limiter.Allow(key) {
		t.Error("request after refill should be allowed")
	}
}

func TestRateLimiterDifferentKeys(t *testing.T) {
	limiter := NewRateLimiter(10, 1)

	// Different keys should have independent limits
	if !limiter.Allow("key1") {
		t.Error("key1 first request should be allowed")
	}
	if !limiter.Allow("key2") {
		t.Error("key2 first request should be allowed")
	}

	// key1 exhausted
	if limiter.Allow("key1") {
		t.Error("key1 second request should be denied")
	}
	// key2 also exhausted
	if limiter.Allow("key2") {
		t.Error("key2 second request should be denied")
	}
}

func TestRateLimiterBurstCapacity(t *testing.T) {
	limiter := NewRateLimiter(5, 20)
	key := "burst-key"

	// All 20 burst requests should succeed
	for i := 0; i < 20; i++ {
		if !limiter.Allow(key) {
			t.Errorf("burst request %d should be allowed", i)
		}
	}

	// 21st should fail
	if limiter.Allow(key) {
		t.Error("request exceeding burst should be denied")
	}
}

func TestRateLimiterTokenRefillRate(t *testing.T) {
	// 50 tokens per second, burst of 5
	limiter := NewRateLimiter(50, 5)
	key := "refill-key"

	// Use all tokens
	for i := 0; i < 5; i++ {
		limiter.Allow(key)
	}

	// Should be denied
	if limiter.Allow(key) {
		t.Error("should be denied after burst exhausted")
	}

	// Wait 50ms = ~2.5 tokens at 50 RPS
	time.Sleep(50 * time.Millisecond)

	// Should allow at least 2 requests
	allowed := 0
	for i := 0; i < 3; i++ {
		if limiter.Allow(key) {
			allowed++
		}
	}

	if allowed < 2 {
		t.Errorf("expected at least 2 requests after refill, got %d", allowed)
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	limiter := NewRateLimiter(100, 100)
	key := "concurrent-key"

	// Run concurrent requests
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			limiter.Allow(key)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	// Should still work after concurrent access
	if !limiter.Allow(key + "-new") {
		t.Error("new key should be allowed after concurrent access")
	}
}

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10, 5)

	if limiter == nil {
		t.Fatal("expected limiter to be created")
	}
	if limiter.rate != 10 {
		t.Errorf("expected rate 10, got %d", limiter.rate)
	}
	if limiter.burst != 5 {
		t.Errorf("expected burst 5, got %d", limiter.burst)
	}
	if limiter.buckets == nil {
		t.Error("buckets map should be initialized")
	}
}

func TestRateLimiterZeroTokens(t *testing.T) {
	limiter := NewRateLimiter(1, 1)
	key := "zero-key"

	// First should succeed
	if !limiter.Allow(key) {
		t.Error("first request should be allowed")
	}

	// Immediately after should fail
	if limiter.Allow(key) {
		t.Error("immediate second request should be denied")
	}

	// Another immediate request should also fail
	if limiter.Allow(key) {
		t.Error("third immediate request should be denied")
	}
}

func TestRateLimiterManyKeys(t *testing.T) {
	limiter := NewRateLimiter(10, 5)

	// Create many different keys
	for i := 0; i < 100; i++ {
		key := string(rune('a' + (i % 26))) + string(rune('0'+(i/26)))
		if !limiter.Allow(key) {
			t.Errorf("first request for key %s should be allowed", key)
		}
	}
}
