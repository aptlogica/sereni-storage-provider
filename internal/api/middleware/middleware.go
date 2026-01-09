package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID middleware sets a X-Request-ID header
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

// Simple in-memory rate limiter per IP (token bucket). Not suitable for multi-instance.
type limiter struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
}

var clients = map[string]*limiter{}
var clientsMu sync.Mutex

func init() {
	go janitor()
}

// janitor periodically removes stale client entries to avoid unbounded memory growth.
func janitor() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		clientsMu.Lock()
		for k, v := range clients {
			v.mu.Lock()
			if v.last.Before(cutoff) {
				delete(clients, k)
			}
			v.mu.Unlock()
		}
		clientsMu.Unlock()
	}
}

func getLimiter(ip string) *limiter {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	l, ok := clients[ip]
	if !ok {
		l = &limiter{tokens: 10, last: time.Now()}
		clients[ip] = l
	}
	return l
}

// RateLimit middleware with simple refill semantics: 10 tokens burst, refill 1 token/sec
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := getLimiter(ip)
		l.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(l.last).Seconds()
		l.tokens += elapsed
		if l.tokens > 10 {
			l.tokens = 10
		}
		l.last = now
		if l.tokens < 1 {
			l.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		l.tokens -= 1
		l.mu.Unlock()
		c.Next()
	}
}
