package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"sereni-storage-provider/internal/api/handlers"
	"sereni-storage-provider/internal/api/routes"
	"sereni-storage-provider/internal/services"

	"github.com/gin-gonic/gin"
)

// MockStorageProvider implements the StorageProvider interface for testing
type MockStorageProvider struct{}

func (m *MockStorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	return "http://mock-url.com/file", nil
}

func (m *MockStorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *MockStorageProvider) Delete(ctx context.Context, objectName string) error {
	return nil
}

func (m *MockStorageProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	return true, nil
}

func (m *MockStorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	return "http://mock-url.com/file", nil
}

func (m *MockStorageProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockStorageProvider) GetSize(ctx context.Context, path string) (int64, bool, error) {
	return 0, false, nil
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock provider
	mockProvider := &MockStorageProvider{}

	// Create real storage service with mock provider
	storageService := services.NewStorageService(mockProvider)

	// Create handlers with real service
	storageHandler := handlers.NewStorageHandler(storageService, 10*1024*1024, []string{"image/jpeg", "image/png"})
	healthHandler := handlers.NewHealthHandler(storageService)

	// Create router
	router := gin.New()
	storagePath := "/tmp/test"

	// Setup routes
	routes.SetupRoutes(router, storageHandler, healthHandler, storagePath)

	// Test cases for different routes
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "swagger route exists",
			method:         "GET",
			path:           "/swagger/index.html",
			expectedStatus: http.StatusOK, // Swagger handler should respond
		},
		{
			name:           "health route exists",
			method:         "GET",
			path:           "/api/v1/health",
			expectedStatus: http.StatusOK, // Health handler should respond
		},
		{
			name:           "upload route accepts POST",
			method:         "POST",
			path:           "/api/v1/storage/upload",
			expectedStatus: http.StatusBadRequest, // Storage handler will return 400 for no file
		},
		{
			name:           "upload route rejects GET",
			method:         "GET",
			path:           "/api/v1/storage/upload",
			expectedStatus: http.StatusNotFound, // Gin returns 404 for method not allowed by default
		},
		{
			name:           "download route accepts GET",
			method:         "GET",
			path:           "/api/v1/storage/download",
			expectedStatus: http.StatusBadRequest, // Storage handler will return 400 for missing path
		},
		{
			name:           "download route rejects POST",
			method:         "POST",
			path:           "/api/v1/storage/download",
			expectedStatus: http.StatusNotFound, // Gin returns 404 for method not allowed by default
		},
		{
			name:           "delete route accepts DELETE",
			method:         "DELETE",
			path:           "/api/v1/storage/delete",
			expectedStatus: http.StatusBadRequest, // Storage handler will return 400 for missing path
		},
		{
			name:           "delete route rejects GET",
			method:         "GET",
			path:           "/api/v1/storage/delete",
			expectedStatus: http.StatusNotFound, // Gin returns 404 for method not allowed by default
		},
		{
			name:           "non-existent route returns 404",
			method:         "GET",
			path:           "/api/v1/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d for %s %s, got %d", tt.expectedStatus, tt.method, tt.path, w.Code)
			}
		})
	}
}
