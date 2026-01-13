package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw := RequestID()
	mw(c)

	if c.Writer.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID header to be set")
	}
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Clear any existing limiters for this IP
	clientsMu.Lock()
	delete(clients, "127.0.0.1")
	clientsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "127.0.0.1")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw := RateLimit()
	mw(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for first request, got %d", w.Code)
	}

	// Make 10 more requests quickly to exhaust tokens
	for i := 0; i < 10; i++ {
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = req
		mw(c)
	}

	// Next request should be rate limited
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	mw(c)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429 for rate limited request, got %d", w.Code)
	}
}

func TestJanitor(t *testing.T) {
	// Add some test clients
	getLimiter("127.0.0.1")
	getLimiter("127.0.0.2")

	// Set one client to be old
	clientsMu.Lock()
	if client, exists := clients["127.0.0.1"]; exists {
		client.last = time.Now().Add(-15 * time.Minute) // Older than 10 minutes
	}
	clientsMu.Unlock()

	// Run janitor with short interval
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	janitorWithInterval(ctx, 10*time.Millisecond)

	// Check that old client was removed
	clientsMu.Lock()
	_, exists := clients["127.0.0.1"]
	clientsMu.Unlock()

	if exists {
		t.Errorf("expected old client to be removed by janitor")
	}

	// Check that new client still exists
	clientsMu.Lock()
	_, exists = clients["127.0.0.2"]
	clientsMu.Unlock()

	if !exists {
		t.Errorf("expected new client to still exist")
	}
}
