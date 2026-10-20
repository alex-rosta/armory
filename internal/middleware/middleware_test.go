package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	now := time.Unix(1000, 0)
	l := NewRateLimiter(1, 3)
	l.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("request %d should be allowed within burst", i)
		}
	}
	ok, retry := l.Allow("a")
	if ok || retry <= 0 {
		t.Fatalf("4th request: ok=%v retry=%v, want denied with retry", ok, retry)
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Fatal("other keys must be unaffected")
	}

	now = now.Add(2 * time.Second)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("tokens should have refilled")
	}

	now = now.Add(time.Hour)
	l.Allow("c")
	if len(l.buckets) != 1 {
		t.Fatalf("idle buckets not cleaned up: %d", len(l.buckets))
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	h := RateLimitMiddleware(NewRateLimiter(0.001, 1), "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	do := func(path string) int {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if c := do("/"); c != http.StatusOK {
		t.Fatalf("first = %d", c)
	}
	if c := do("/"); c != http.StatusTooManyRequests {
		t.Fatalf("second = %d, want 429", c)
	}
	if c := do("/assets/css/styles.css"); c != http.StatusOK {
		t.Fatalf("assets = %d, want 200 (not rate limited)", c)
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	req.Header.Set("Fly-Client-IP", "203.0.113.5")
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")

	if got := ClientIP(req, ""); got != "192.0.2.1" {
		t.Errorf("no header configured: %q", got)
	}
	if got := ClientIP(req, "Fly-Client-IP"); got != "203.0.113.5" {
		t.Errorf("Fly-Client-IP: %q", got)
	}
	if got := ClientIP(req, "X-Forwarded-For"); got != "203.0.113.9" {
		t.Errorf("last XFF entry should be used: %q", got)
	}
	req.Header.Set("Fly-Client-IP", "not-an-ip")
	if got := ClientIP(req, "Fly-Client-IP"); got != "192.0.2.1" {
		t.Errorf("invalid header should fall back: %q", got)
	}
}

func TestAllowReadOnlyMiddleware(t *testing.T) {
	h := AllowReadOnlyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	for method, want := range map[string]int{
		http.MethodGet: 200, http.MethodHead: 200,
		http.MethodPost: 405, http.MethodPut: 405, http.MethodDelete: 405, http.MethodTrace: 405,
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, "/", nil))
		if rec.Code != want {
			t.Errorf("%s = %d, want %d", method, rec.Code, want)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, name := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Content-Security-Policy", "Strict-Transport-Security", "Permissions-Policy"} {
		if rec.Header().Get(name) == "" {
			t.Errorf("missing header %s", name)
		}
	}
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Error("CSP missing frame-ancestors")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("boom") })).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d", rec.Code)
	}
}
