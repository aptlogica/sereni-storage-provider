package middleware

import (
	"context"
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
	Last   time.Time
}

var Clients = map[string]*limiter{}
var ClientsMu sync.Mutex

func init() {
	go janitor()
}

// janitor periodically removes stale client entries to avoid unbounded memory growth.
func janitor() {
	JanitorWithInterval(context.Background(), 5*time.Minute)
}

// JanitorWithInterval allows testing with custom interval and context
func JanitorWithInterval(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-10 * time.Minute)
			ClientsMu.Lock()
			for k, v := range Clients {
				v.mu.Lock()
				if v.Last.Before(cutoff) {
					delete(Clients, k)
				}
				v.mu.Unlock()
			}
			ClientsMu.Unlock()
		}
	}
}

func GetLimiter(ip string) *limiter {
	ClientsMu.Lock()
	defer ClientsMu.Unlock()
	l, ok := Clients[ip]
	if !ok {
		l = &limiter{tokens: 20, Last: time.Now()}
		Clients[ip] = l
	}
	return l
}

// RateLimit middleware with simple refill semantics: 10 tokens burst, refill 1 token/sec
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := GetLimiter(ip)
		l.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(l.Last).Seconds()
		l.tokens += elapsed
		if l.tokens > 10 {
			l.tokens = 10
		}
		l.Last = now
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
