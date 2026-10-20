package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

// RateLimiter is an in-memory per-key token bucket limiter
type RateLimiter struct {
	rate  float64 // tokens added per second
	burst float64

	mu          sync.Mutex
	buckets     map[string]*bucket
	lastCleanup time.Time
	now         func() time.Time
}

// NewRateLimiter creates a limiter allowing ratePerSecond sustained requests per key with the given burst
func NewRateLimiter(ratePerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:    ratePerSecond,
		burst:   float64(burst),
		buckets: make(map[string]*bucket),
		now:     time.Now,
	}
}

// Allow reports whether a request for key may proceed; when it may not, retryAfter is how long to wait
func (l *RateLimiter) Allow(key string) (ok bool, retryAfter time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.cleanup(now)

	b, exists := l.buckets[key]
	if !exists {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}

	b.tokens = math.Min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now

	if b.tokens < 1 {
		wait := time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
		return false, wait
	}
	b.tokens--
	return true, 0
}

// cleanup drops buckets that have fully refilled, which are equivalent to absent ones; callers hold l.mu
func (l *RateLimiter) cleanup(now time.Time) {
	if now.Sub(l.lastCleanup) < time.Minute {
		return
	}
	l.lastCleanup = now
	refill := time.Duration(l.burst / l.rate * float64(time.Second))
	for k, b := range l.buckets {
		if now.Sub(b.last) > refill {
			delete(l.buckets, k)
		}
	}
}

// RateLimitMiddleware limits requests per client IP; static assets are not counted
func RateLimitMiddleware(limiter *RateLimiter, clientIPHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				next.ServeHTTP(w, r)
				return
			}
			if ok, retryAfter := limiter.Allow(ClientIP(r, clientIPHeader)); !ok {
				w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
