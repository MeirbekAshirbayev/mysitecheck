package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Rate Limiter ---

// IPRateLimiter using standard library (Fixed Window Counter)
type IPRateLimiter struct {
	ips    map[string]int
	expiry map[string]time.Time
	mu     sync.Mutex
	limit  int
	window time.Duration
}

var limiter = &IPRateLimiter{
	ips:    make(map[string]int),
	expiry: make(map[string]time.Time),
	limit:  50,               // 50 requests
	window: 10 * time.Second, // per 10 seconds
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		// Reset if window expired
		if time.Now().After(limiter.expiry[ip]) {
			limiter.ips[ip] = 0
			limiter.expiry[ip] = time.Now().Add(limiter.window)
		}

		limiter.ips[ip]++

		if limiter.ips[ip] > limiter.limit {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}

// --- Security Headers ---

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protect against XSS
		c.Header("X-XSS-Protection", "1; mode=block")
		// Protect against Clickjacking
		c.Header("X-Frame-Options", "SAMEORIGIN")
		// Prevent MIME-sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// --- CSRF ---

// CSRF Protection for Admin Routes
func CsrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			referer := c.Request.Referer()
			origin := c.Request.Header.Get("Origin")
			host := c.Request.Host

			// If Referer is present, check it
			if referer != "" {
				if !strings.Contains(referer, host) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid Referer"})
					return
				}
			}
			// If Origin is present, check it (often sent by browsers on POST)
			if origin != "" {
				if !strings.Contains(origin, host) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid Origin"})
					return
				}
			}
		}
		c.Next()
	}
}
