// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middlewarePkg "sereni-storage-provider/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw := middlewarePkg.RequestID()
	mw(c)

	if c.Writer.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID header to be set")
	}
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Clear any existing limiters for this IP
	middlewarePkg.ClientsMu.Lock()
	delete(middlewarePkg.Clients, "127.0.0.1")
	middlewarePkg.ClientsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "127.0.0.1")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mw := middlewarePkg.RateLimit()
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
	// Add some test middleware.Clients
	middlewarePkg.GetLimiter("127.0.0.1")
	middlewarePkg.GetLimiter("127.0.0.2")

	// Set one client to be old
	middlewarePkg.ClientsMu.Lock()
	if client, exists := middlewarePkg.Clients["127.0.0.1"]; exists {
		client.Last = time.Now().Add(-15 * time.Minute) // Older than 10 minutes
	}
	middlewarePkg.ClientsMu.Unlock()

	// Run janitor with short interval
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	middlewarePkg.JanitorWithInterval(ctx, 10*time.Millisecond)

	// Check that old client was removed
	middlewarePkg.ClientsMu.Lock()
	_, exists := middlewarePkg.Clients["127.0.0.1"]
	middlewarePkg.ClientsMu.Unlock()

	if exists {
		t.Errorf("expected old client to be removed by janitor")
	}

	// Check that new client still exists
	middlewarePkg.ClientsMu.Lock()
	_, exists = middlewarePkg.Clients["127.0.0.2"]
	middlewarePkg.ClientsMu.Unlock()

	if !exists {
		t.Errorf("expected new client to still exist")
	}
}
