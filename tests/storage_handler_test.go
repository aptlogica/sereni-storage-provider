// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
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
	getSizeFn    func(ctx context.Context, path string) (int64, bool, error)
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

func (m *mockStorageProviderHandler) GetSize(ctx context.Context, path string) (int64, bool, error) {
	if m.getSizeFn != nil {
		return m.getSizeFn(ctx, path)
	}
	return 0, false, nil
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

func TestStorageHandler_GetConsumption(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		mockSetup      func(*mockStorageProviderHandler)
		expectedStatus int
		checkBody      func(*testing.T, string)
	}{
		{
			name: "successful file consumption",
			path: "uploads/test.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 1024, false, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"size_bytes":1024`) {
					t.Errorf("expected size_bytes 1024 in body: %s", body)
				}
				if !strings.Contains(body, `"is_directory":false`) {
					t.Errorf("expected is_directory false in body: %s", body)
				}
				if !strings.Contains(body, `"size_human":"1.00 KB"`) {
					t.Errorf("expected size_human '1.00 KB' in body: %s", body)
				}
			},
		},
		{
			name: "successful directory consumption",
			path: "uploads/folder",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 5242880, true, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"size_bytes":5242880`) {
					t.Errorf("expected size_bytes 5242880 in body: %s", body)
				}
				if !strings.Contains(body, `"is_directory":true`) {
					t.Errorf("expected is_directory true in body: %s", body)
				}
				if !strings.Contains(body, `"size_human":"5.00 MB"`) {
					t.Errorf("expected size_human '5.00 MB' in body: %s", body)
				}
			},
		},
		{
			name: "small file in bytes",
			path: "test.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 512, false, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"size_human":"512 B"`) {
					t.Errorf("expected size_human '512 B' in body: %s", body)
				}
			},
		},
		{
			name: "large file in GB",
			path: "large.zip",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 2147483648, false, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"size_human":"2.00 GB"`) {
					t.Errorf("expected size_human '2.00 GB' in body: %s", body)
				}
			},
		},
		{
			name: "zero size file",
			path: "empty.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 0, false, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"size_human":"0 B"`) {
					t.Errorf("expected size_human '0 B' in body: %s", body)
				}
			},
		},
		{
			name: "very large file in TB",
			path: "huge.dat",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 1099511627776, false, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"size_human":"1.00 TB"`) {
					t.Errorf("expected size_human '1.00 TB' in body: %s", body)
				}
			},
		},
		{
			name:           "missing path parameter",
			path:           "",
			mockSetup:      func(m *mockStorageProviderHandler) {},
			expectedStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"error":"Path parameter is required"`) {
					t.Errorf("expected error 'Path parameter is required' in body: %s", body)
				}
			},
		},
		{
			name: "file not found",
			path: "nonexistent.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 0, false, app_errors.FileNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"error":"file or directory not found"`) {
					t.Errorf("expected error 'file or directory not found' in body: %s", body)
				}
			},
		},
		{
			name: "invalid path",
			path: "../etc/passwd",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					// NormalizePath will return empty string for path traversal
					// So the path will be empty when it reaches the provider
					if path == "" {
						return 0, false, nil
					}
					return 0, false, app_errors.ErrInvalidPath
				}
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				// Path traversal is normalized to empty, returns 0 size
				if !strings.Contains(body, `"size_bytes":0`) {
					t.Errorf("expected size_bytes 0 in body: %s", body)
				}
			},
		},
		{
			name: "internal server error",
			path: "test.txt",
			mockSetup: func(m *mockStorageProviderHandler) {
				m.getSizeFn = func(ctx context.Context, path string) (int64, bool, error) {
					return 0, false, app_errors.ErrGenericStorageFailed
				}
			},
			expectedStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, `"error":"failed to get size information"`) {
					t.Errorf("expected error 'failed to get size information' in body: %s", body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockStorageProviderHandler{}
			tt.mockSetup(provider)
			service := services.NewStorageService(provider)
			handler := handlers.NewStorageHandler(service, 1024, nil)

			req := httptest.NewRequest(http.MethodGet, "/consumption?path="+tt.path, nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.GetConsumption(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			tt.checkBody(t, w.Body.String())
		})
	}
}
