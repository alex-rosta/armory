package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
)

const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' 'unsafe-inline' https://unpkg.com; " +
	"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; " +
	"font-src https://cdn.jsdelivr.net; " +
	"img-src 'self' data: https://render.worldofwarcraft.com; " +
	"connect-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

// ClientIP returns the caller's IP. When header is set it is trusted (it must be
// set by a reverse proxy that overwrites client-supplied values); for a
// comma-separated list the last entry, added by the nearest proxy, is used.
func ClientIP(r *http.Request, header string) string {
	if header != "" {
		if v := r.Header.Get(header); v != "" {
			if i := strings.LastIndex(v, ","); i >= 0 {
				v = v[i+1:]
			}
			if ip := net.ParseIP(strings.TrimSpace(v)); ip != nil {
				return ip.String()
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// LoggingMiddleware logs HTTP requests; the path is quoted to prevent log injection
func LoggingMiddleware(clientIPHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %q", ClientIP(r, clientIPHeader), r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

// AllowReadOnlyMiddleware rejects everything except GET and HEAD
func AllowReadOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware sets common security response headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("Strict-Transport-Security", "max-age=31536000")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if err == http.ErrAbortHandler {
					panic(err)
				}
				log.Printf("recovered from panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
