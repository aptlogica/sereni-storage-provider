package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sereni-storage-provider/internal/api/handlers"
	"sereni-storage-provider/internal/services"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Mock provider
type mockHealthProvider struct {
	healthCheckFn func(ctx context.Context) error
}

func (m *mockHealthProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	return "", nil
}

func (m *mockHealthProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockHealthProvider) Delete(ctx context.Context, objectName string) error {
	return nil
}

func (m *mockHealthProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	return false, nil
}

func (m *mockHealthProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	return "", nil
}

func (m *mockHealthProvider) HealthCheck(ctx context.Context) error {
	return m.healthCheckFn(ctx)
}

func TestNewHealthHandler(t *testing.T) {
	provider := &mockHealthProvider{}
	service := services.NewStorageService(provider)
	handler := handlers.NewHealthHandler(service)
	if handler == nil {
		t.Fatalf("expected handler, got nil")
	}
}

func TestHealthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockSetup      func(*mockHealthProvider)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "healthy",
			mockSetup: func(m *mockHealthProvider) {
				m.healthCheckFn = func(ctx context.Context) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"healthy"}`,
		},
		{
			name: "unhealthy",
			mockSetup: func(m *mockHealthProvider) {
				m.healthCheckFn = func(ctx context.Context) error {
					return io.EOF
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"EOF","status":"error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockHealthProvider{}
			tt.mockSetup(provider)
			service := services.NewStorageService(provider)
			handler := handlers.NewHealthHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.Health(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if strings.TrimSpace(w.Body.String()) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
