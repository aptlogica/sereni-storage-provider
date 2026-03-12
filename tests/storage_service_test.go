// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package tests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	servicesPkg "sereni-storage-provider/internal/services"
	"strings"
	"testing"
)

// Mock StorageProvider
type mockStorageProviderService struct {
	UploadFn      func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	DownloadFn    func(ctx context.Context, objectName string) (io.ReadCloser, error)
	DeleteFn      func(ctx context.Context, objectName string) error
	ExistsFn      func(ctx context.Context, objectName string) (bool, error)
	GetURLFn      func(ctx context.Context, objectName string) (string, error)
	GetSizeFn     func(ctx context.Context, path string) (int64, bool, error)
	HealthCheckFn func(ctx context.Context) error
}

func (m *mockStorageProviderService) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	return m.UploadFn(ctx, objectName, reader, size, contentType)
}

func (m *mockStorageProviderService) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return m.DownloadFn(ctx, objectName)
}

func (m *mockStorageProviderService) Delete(ctx context.Context, objectName string) error {
	return m.DeleteFn(ctx, objectName)
}

func (m *mockStorageProviderService) Exists(ctx context.Context, objectName string) (bool, error) {
	return m.ExistsFn(ctx, objectName)
}

func (m *mockStorageProviderService) GetURL(ctx context.Context, objectName string) (string, error) {
	return m.GetURLFn(ctx, objectName)
}

func (m *mockStorageProviderService) HealthCheck(ctx context.Context) error {
	return m.HealthCheckFn(ctx)
}

func (m *mockStorageProviderService) GetSize(ctx context.Context, path string) (int64, bool, error) {
	if m.GetSizeFn != nil {
		return m.GetSizeFn(ctx, path)
	}
	return 0, false, nil
}

func TestNewStorageService(t *testing.T) {
	provider := &mockStorageProviderService{}
	service := servicesPkg.NewStorageService(provider)
	if service == nil {
		t.Fatalf("expected service, got nil")
	}
	if provider != provider {
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
			result := servicesPkg.NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
func TestStorageService_GetFile(t *testing.T) {
	provider := &mockStorageProviderService{
		DownloadFn: func(ctx context.Context, objectName string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content")), nil
		},
	}
	service := servicesPkg.NewStorageService(provider)

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
	provider := &mockStorageProviderService{
		DeleteFn: func(ctx context.Context, objectName string) error {
			return nil
		},
	}
	service := servicesPkg.NewStorageService(provider)

	err := service.DeleteFile(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageService_FileExists(t *testing.T) {
	provider := &mockStorageProviderService{
		ExistsFn: func(ctx context.Context, objectName string) (bool, error) {
			return true, nil
		},
	}
	service := servicesPkg.NewStorageService(provider)

	exists, err := service.FileExists(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("expected true, got false")
	}
}

func TestStorageService_GetFileURL(t *testing.T) {
	provider := &mockStorageProviderService{
		GetURLFn: func(ctx context.Context, objectName string) (string, error) {
			return "http://example.com/test.txt", nil
		},
	}
	service := servicesPkg.NewStorageService(provider)

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
		name         string
		fileName     string
		path         string
		setupFile    func() *multipart.FileHeader
		mockUpload   func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
		expectError  bool
		expectedURL  string
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
			provider := &mockStorageProviderService{
				UploadFn: tt.mockUpload,
			}
			service := servicesPkg.NewStorageService(provider)

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
	provider := &mockStorageProviderService{
		HealthCheckFn: func(ctx context.Context) error {
			return nil
		},
	}
	service := servicesPkg.NewStorageService(provider)

	err := service.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageService_GetSize(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		mockFunc      func(ctx context.Context, path string) (int64, bool, error)
		expectedSize  int64
		expectedIsDir bool
		expectError   bool
	}{
		{
			name: "successful file size",
			path: "uploads/test.txt",
			mockFunc: func(ctx context.Context, path string) (int64, bool, error) {
				return 1024, false, nil
			},
			expectedSize:  1024,
			expectedIsDir: false,
			expectError:   false,
		},
		{
			name: "successful directory size",
			path: "uploads/folder",
			mockFunc: func(ctx context.Context, path string) (int64, bool, error) {
				return 5242880, true, nil
			},
			expectedSize:  5242880,
			expectedIsDir: true,
			expectError:   false,
		},
		{
			name: "empty path",
			path: "",
			mockFunc: func(ctx context.Context, path string) (int64, bool, error) {
				return 0, false, nil
			},
			expectedSize:  0,
			expectedIsDir: false,
			expectError:   false,
		},
		{
			name: "path with leading slash",
			path: "/uploads/test.txt",
			mockFunc: func(ctx context.Context, path string) (int64, bool, error) {
				if path != "uploads/test.txt" {
					t.Errorf("expected normalized path 'uploads/test.txt', got '%s'", path)
				}
				return 2048, false, nil
			},
			expectedSize:  2048,
			expectedIsDir: false,
			expectError:   false,
		},
		{
			name: "error from provider",
			path: "test.txt",
			mockFunc: func(ctx context.Context, path string) (int64, bool, error) {
				return 0, false, errors.New("provider error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockStorageProviderService{
				GetSizeFn: tt.mockFunc,
			}
			service := servicesPkg.NewStorageService(provider)

			size, isDir, err := service.GetSize(context.Background(), tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if size != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, size)
			}

			if isDir != tt.expectedIsDir {
				t.Errorf("expected isDir %v, got %v", tt.expectedIsDir, isDir)
			}
		})
	}
}
