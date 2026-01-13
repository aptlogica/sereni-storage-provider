package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"
)

// Mock StorageProvider
type mockStorageProvider struct {
	UploadFn      func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	DownloadFn    func(ctx context.Context, objectName string) (io.ReadCloser, error)
	DeleteFn      func(ctx context.Context, objectName string) error
	ExistsFn      func(ctx context.Context, objectName string) (bool, error)
	GetURLFn      func(ctx context.Context, objectName string) (string, error)
	HealthCheckFn func(ctx context.Context) error
}

func (m *mockStorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	return m.UploadFn(ctx, objectName, reader, size, contentType)
}

func (m *mockStorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return m.DownloadFn(ctx, objectName)
}

func (m *mockStorageProvider) Delete(ctx context.Context, objectName string) error {
	return m.DeleteFn(ctx, objectName)
}

func (m *mockStorageProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	return m.ExistsFn(ctx, objectName)
}

func (m *mockStorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	return m.GetURLFn(ctx, objectName)
}

func (m *mockStorageProvider) HealthCheck(ctx context.Context) error {
	return m.HealthCheckFn(ctx)
}

func TestNewStorageService(t *testing.T) {
	provider := &mockStorageProvider{}
	service := NewStorageService(provider)
	if service == nil {
		t.Fatalf("expected service, got nil")
	}
	if service.provider != provider {
		t.Fatalf("expected provider to be set")
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "simple",
			input:    "file.txt",
			expected: "file.txt",
		},
		{
			name:     "with slash",
			input:    "path/file.txt",
			expected: "path/file.txt",
		},
		{
			name:     "multiple slashes",
			input:    "path//file.txt",
			expected: "path/file.txt",
		},
		{
			name:     "backslashes",
			input:    "path\\file.txt",
			expected: "path/file.txt",
		},
		{
			name:     "dot",
			input:    ".",
			expected: "",
		},
		{
			name:     "dot dot",
			input:    "..",
			expected: "",
		},
		{
			name:     "leading slash",
			input:    "/file.txt",
			expected: "file.txt",
		},
		{
			name:     "path traversal",
			input:    "../file.txt",
			expected: "",
		},
		{
			name:     "complex",
			input:    "uploads/../file.txt",
			expected: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
func TestStorageService_GetFile(t *testing.T) {
	provider := &mockStorageProvider{
		DownloadFn: func(ctx context.Context, objectName string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content")), nil
		},
	}
	service := NewStorageService(provider)

	reader, err := service.GetFile(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("got %v, want content", string(data))
	}
}

func TestStorageService_DeleteFile(t *testing.T) {
	provider := &mockStorageProvider{
		DeleteFn: func(ctx context.Context, objectName string) error {
			return nil
		},
	}
	service := NewStorageService(provider)

	err := service.DeleteFile(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageService_FileExists(t *testing.T) {
	provider := &mockStorageProvider{
		ExistsFn: func(ctx context.Context, objectName string) (bool, error) {
			return true, nil
		},
	}
	service := NewStorageService(provider)

	exists, err := service.FileExists(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("expected true, got false")
	}
}

func TestStorageService_GetFileURL(t *testing.T) {
	provider := &mockStorageProvider{
		GetURLFn: func(ctx context.Context, objectName string) (string, error) {
			return "http://example.com/test.txt", nil
		},
	}
	service := NewStorageService(provider)

	url, err := service.GetFileURL(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "http://example.com/test.txt" {
		t.Errorf("got %v, want http://example.com/test.txt", url)
	}
}

func TestStorageService_UploadFile(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		path        string
		setupFile   func() *multipart.FileHeader
		mockUpload  func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
		expectError bool
		expectedURL string
		expectedName string
	}{
		{
			name:     "successful upload with directory path",
			fileName: "test.txt",
			path:     "uploads/",
			setupFile: func() *multipart.FileHeader {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.txt")
				part.Write([]byte("test content"))
				writer.Close()

				reader := multipart.NewReader(body, writer.Boundary())
				form, _ := reader.ReadForm(32 << 20)
				return form.File["file"][0]
			},
			mockUpload: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
				return "http://example.com/uploads/test.txt", nil
			},
			expectError:  false,
			expectedURL:  "http://example.com/uploads/test.txt",
			expectedName: "uploads/test.txt",
		},
		{
			name:     "successful upload with filename only",
			fileName: "test.txt",
			path:     "",
			setupFile: func() *multipart.FileHeader {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.txt")
				part.Write([]byte("test content"))
				writer.Close()

				reader := multipart.NewReader(body, writer.Boundary())
				form, _ := reader.ReadForm(32 << 20)
				return form.File["file"][0]
			},
			mockUpload: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
				return "http://example.com/test.txt", nil
			},
			expectError:  false,
			expectedURL:  "http://example.com/test.txt",
			expectedName: "test.txt",
		},
		{
			name:     "upload error",
			fileName: "test.txt",
			path:     "uploads/",
			setupFile: func() *multipart.FileHeader {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.txt")
				part.Write([]byte("test content"))
				writer.Close()

				reader := multipart.NewReader(body, writer.Boundary())
				form, _ := reader.ReadForm(32 << 20)
				return form.File["file"][0]
			},
			mockUpload: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
				return "", errors.New("upload failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := tt.setupFile()
			provider := &mockStorageProvider{
				UploadFn: tt.mockUpload,
			}
			service := NewStorageService(provider)

			response, err := service.UploadFile(context.Background(), file, tt.path)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if response.URL != tt.expectedURL {
				t.Errorf("expected URL '%s', got '%s'", tt.expectedURL, response.URL)
			}
			if response.ObjectName != tt.expectedName {
				t.Errorf("expected ObjectName '%s', got '%s'", tt.expectedName, response.ObjectName)
			}
		})
	}
}

func TestStorageService_HealthCheck(t *testing.T) {
	provider := &mockStorageProvider{
		HealthCheckFn: func(ctx context.Context) error {
			return nil
		},
	}
	service := NewStorageService(provider)

	err := service.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
