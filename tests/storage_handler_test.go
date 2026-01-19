package tests

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sereni-storage-provider/internal/api/handlers"
	app_errors "sereni-storage-provider/internal/app-errors"
	"sereni-storage-provider/internal/services"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Mock provider
type mockStorageProviderHandler struct {
	uploadFn     func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	getFileFn    func(ctx context.Context, path string) (io.ReadCloser, error)
	deleteFileFn func(ctx context.Context, path string) error
}

func (m *mockStorageProviderHandler) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	return m.uploadFn(ctx, objectName, reader, size, contentType)
}

func (m *mockStorageProviderHandler) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return m.getFileFn(ctx, objectName)
}

func (m *mockStorageProviderHandler) Delete(ctx context.Context, objectName string) error {
	return m.deleteFileFn(ctx, objectName)
}

func (m *mockStorageProviderHandler) Exists(ctx context.Context, objectName string) (bool, error) {
	return false, nil
}

func (m *mockStorageProviderHandler) GetURL(ctx context.Context, objectName string) (string, error) {
	return "", nil
}

func (m *mockStorageProviderHandler) HealthCheck(ctx context.Context) error {
	return nil
}

func TestNewStorageHandler(t *testing.T) {
	provider := &mockStorageProviderHandler{}
	service := services.NewStorageService(provider)
	handler := handlers.NewStorageHandler(service, 10<<20, []string{"text/plain"})
	if handler == nil {
		t.Fatalf("expected handler, got nil")
	}
}

func TestStorageHandler_Upload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		fileContent    string
		fileName       string
		contentType    string
		path           string
		maxUploadSize  int64
		allowedTypes   []string
		mockSetup      func(*mockStorageProviderHandler)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:          "successful upload",
			fileContent:   "test content",
			fileName:      "test.txt",
			contentType:   "text/plain",
			path:          "uploads",
			maxUploadSize: 100,
			allowedTypes:  []string{"text/plain"},
			mockSetup: func(m *mockStorageProviderHandler) {
				m.uploadFn = func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
					return "http://example.com/uploads/test.txt", nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"File uploaded successfully","path":"uploads/test.txt","url":"http://example.com/uploads/test.txt"}`,
		},
		{
			name:           "no file",
			fileContent:    "",
			fileName:       "",
			contentType:    "",
			path:           "",
			maxUploadSize:  100,
			allowedTypes:   []string{},
			mockSetup:      func(m *mockStorageProviderHandler) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"No file uploaded"}`,
		},
		{
			name:           "file too large",
			fileContent:    strings.Repeat("a", 200),
			fileName:       "large.txt",
			contentType:    "text/plain",
			path:           "uploads",
			maxUploadSize:  100,
			allowedTypes:   []string{"text/plain"},
			mockSetup:      func(m *mockStorageProviderHandler) {},
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectedBody:   `{"error":"file too large"}`,
		},
		{
			name:           "invalid content type",
			fileContent:    "content",
			fileName:       "test.txt",
			contentType:    "text/plain",
			path:           "uploads",
			maxUploadSize:  100,
			allowedTypes:   []string{"image/png"},
			mockSetup:      func(m *mockStorageProviderHandler) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"detected":"text/plain","error":"invalid content type"}`,
		},
		{
			name:          "wildcard content type match",
			fileContent:   "test content",
			fileName:      "test.txt",
			contentType:   "text/plain",
			path:          "uploads",
			maxUploadSize: 100,
			allowedTypes:  []string{"text/*"},
			mockSetup: func(m *mockStorageProviderHandler) {
				m.uploadFn = func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
					return "http://example.com/uploads/test.txt", nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"File uploaded successfully","path":"uploads/test.txt","url":"http://example.com/uploads/test.txt"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockStorageProviderHandler{}
			tt.mockSetup(provider)
			service := services.NewStorageService(provider)
			handler := handlers.NewStorageHandler(service, tt.maxUploadSize, tt.allowedTypes)

			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			if tt.fileName != "" {
				part, _ := writer.CreateFormFile("file", tt.fileName)
				part.Write([]byte(tt.fileContent))
			}
			if tt.path != "" {
				writer.WriteField("path", tt.path)
			}
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.Upload(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if strings.TrimSpace(w.Body.String()) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestStorageHandler_Download(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		mockSetup      func(*mockStorageProviderHandler)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful download",
			path: "test.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getFileFn = func(ctx context.Context, path string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("file content")), nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "file content",
		},
		{
			name: "file not found",
			path: "missing.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getFileFn = func(ctx context.Context, path string) (io.ReadCloser, error) {
					return nil, app_errors.FileNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"file not found"}`,
		},
		{
			name: "invalid path",
			path: "../invalid",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getFileFn = func(ctx context.Context, path string) (io.ReadCloser, error) {
					return nil, app_errors.ErrInvalidPath
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid path"}`,
		},
		{
			name:           "no path",
			path:           "",
			mockSetup:      func(m *mockStorageProviderHandler) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Path parameter is required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockStorageProviderHandler{}
			tt.mockSetup(provider)
			service := services.NewStorageService(provider)
			handler := handlers.NewStorageHandler(service, 1024, nil)

			req := httptest.NewRequest(http.MethodGet, "/download?path="+tt.path, nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.Download(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				if strings.TrimSpace(w.Body.String()) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

func TestStorageHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		mockSetup      func(*mockStorageProviderHandler)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful delete",
			path: "test.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.deleteFileFn = func(ctx context.Context, path string) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"File deleted successfully"}`,
		},
		{
			name: "file not found",
			path: "missing.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.deleteFileFn = func(ctx context.Context, path string) error {
					return app_errors.FileNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"file not found"}`,
		},
		{
			name: "invalid path",
			path: "../invalid",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.deleteFileFn = func(ctx context.Context, path string) error {
					return app_errors.ErrInvalidPath
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid path"}`,
		},
		{
			name:           "no path",
			path:           "",
			mockSetup:      func(m *mockStorageProviderHandler) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Path parameter is required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockStorageProviderHandler{}
			tt.mockSetup(provider)
			service := services.NewStorageService(provider)
			handler := handlers.NewStorageHandler(service, 1024, nil)

			req := httptest.NewRequest(http.MethodDelete, "/delete?path="+tt.path, nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.Delete(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if strings.TrimSpace(w.Body.String()) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
