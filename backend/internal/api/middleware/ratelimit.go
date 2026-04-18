package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"backend/internal/audit"
)

// ipLimiter holds the token-bucket rate limiter and last-seen time for one IP.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// loginRateLimiter manages per-IP token-bucket limiters.
type loginRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*ipLimiter

	// Policy: max 5 attempts per minute per IP (burst=5, refill=5/min)
	rateLimit rate.Limit
	burst     int
}

func newLoginRateLimiter() *loginRateLimiter {
	rl := &loginRateLimiter{
		visitors:  make(map[string]*ipLimiter),
		rateLimit: rate.Every(12 * time.Second), // 5 tokens per minute (1 every 12s)
		burst:     5,
	}
	go rl.cleanupLoop()
	return rl
}

// getLimiter returns (or creates) the per-IP limiter.
func (rl *loginRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &ipLimiter{
			limiter: rate.NewLimiter(rl.rateLimit, rl.burst),
		}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupLoop removes IPs that haven't been seen in the last 5 minutes.
// This prevents the map from growing unbounded on high-traffic servers.
func (rl *loginRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// package-level singleton — initialised once at server startup.
var loginLimiter = newLoginRateLimiter()

// LoginRateLimitMiddleware enforces a maximum of 5 login attempts per minute per IP.
// Exceeding the limit results in HTTP 429 and a Zap-logged security alert.
func LoginRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the real client IP (respects X-Forwarded-For from reverse proxies)
		ip := clientIP(r)

		if !loginLimiter.getLimiter(ip).Allow() {
			// Log security event — useful for detecting brute-force campaigns
			audit.Logger.Warn("Rate limit exceeded on login endpoint",
				zap.String("ip", ip),
				zap.String("path", r.URL.Path),
				zap.String("user_agent", r.UserAgent()),
			)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Too many login attempts. Please wait before trying again.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the originating IP address from the request.
// It checks X-Real-IP and X-Forwarded-For headers for requests behind proxies.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can be a comma-separated list; first entry is the client
		if host, _, err := net.SplitHostPort(ip); err == nil {
			return host
		}
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
