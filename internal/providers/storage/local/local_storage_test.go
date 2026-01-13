package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sereni-storage-provider/internal/config"
	"strings"
	"testing"
)

func TestNewLocalStorageProvider(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.StorageDevConfig
		serverCfg   config.ServerConfig
		expectError bool
	}{
		{
			name: "valid config",
			cfg: config.StorageDevConfig{
				Path: t.TempDir(),
			},
			serverCfg: config.ServerConfig{
				Scheme: "http",
				Host:   "localhost",
				Port:   "8080",
			},
			expectError: false,
		},
		{
			name: "empty path",
			cfg: config.StorageDevConfig{
				Path: "",
			},
			serverCfg:   config.ServerConfig{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewLocalStorageProvider(&tt.cfg, &tt.serverCfg)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if provider == nil {
				t.Fatalf("expected provider, got nil")
			}
		})
	}
}

func TestLocalStorageProvider_Upload(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.StorageDevConfig{Path: tempDir}
	serverCfg := config.ServerConfig{Scheme: "http", Host: "localhost", Port: "8080"}
	provider, err := NewLocalStorageProvider(&cfg, &serverCfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	tests := []struct {
		name        string
		objectName  string
		content     string
		size        int64
		contentType string
		expectError bool
	}{
		{
			name:        "valid upload",
			objectName:  "test.txt",
			content:     "hello world",
			size:        11,
			contentType: "text/plain",
			expectError: false,
		},
		{
			name:        "upload with subdir",
			objectName:  "subdir/test.txt",
			content:     "content",
			size:        7,
			contentType: "text/plain",
			expectError: false,
		},
		{
			name:        "invalid object name ending with slash",
			objectName:  "test/",
			content:     "content",
			size:        7,
			contentType: "text/plain",
			expectError: true,
		},
		{
			name:        "upload with unknown size",
			objectName:  "unknown_size.txt",
			content:     "unknown",
			size:        -1,
			contentType: "text/plain",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			url, err := provider.Upload(context.Background(), tt.objectName, reader, tt.size, tt.contentType)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedURL := "http://localhost:8080/uploads/" + strings.ReplaceAll(tt.objectName, "\\", "/")
			if url != expectedURL {
				t.Errorf("expected URL %v, got %v", expectedURL, url)
			}

			// Check file exists
			fullPath := filepath.Join(tempDir, tt.objectName)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("file not created: %v", fullPath)
			}

			// Check content
			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			if string(data) != tt.content {
				t.Errorf("file content mismatch: got %v, want %v", string(data), tt.content)
			}
		})
	}
}

func TestLocalStorageProvider_Download(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.StorageDevConfig{Path: tempDir}
	serverCfg := config.ServerConfig{Scheme: "http", Host: "localhost", Port: "8080"}
	provider, err := NewLocalStorageProvider(&cfg, &serverCfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Upload first
	content := "download content"
	_, err = provider.Upload(context.Background(), "download.txt", strings.NewReader(content), int64(len(content)), "text/plain")
	if err != nil {
		t.Fatalf("failed to upload: %v", err)
	}

	tests := []struct {
		name        string
		objectName  string
		expectError bool
	}{
		{
			name:        "success",
			objectName:  "download.txt",
			expectError: false,
		},
		{
			name:        "file not found",
			objectName:  "nonexistent.txt",
			expectError: true,
		},
		{
			name:        "invalid path",
			objectName:  "../outside.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := provider.Download(context.Background(), tt.objectName)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer reader.Close()

			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}
			if string(data) != content {
				t.Errorf("got %v, want %v", string(data), content)
			}
		})
	}
}

func TestLocalStorageProvider_Delete(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.StorageDevConfig{Path: tempDir}
	serverCfg := config.ServerConfig{Scheme: "http", Host: "localhost", Port: "8080"}
	provider, err := NewLocalStorageProvider(&cfg, &serverCfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "delete.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a test directory
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	tests := []struct {
		name        string
		object      string
		expectError bool
	}{
		{
			name:        "delete existing file",
			object:      "delete.txt",
			expectError: false,
		},
		{
			name:        "delete non-existing file",
			object:      "nonexistent.txt",
			expectError: true,
		},
		{
			name:        "delete directory",
			object:      "testdir",
			expectError: true,
		},
		{
			name:        "delete with path traversal",
			object:      "../outside.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Delete(context.Background(), tt.object)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLocalStorageProvider_Exists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.StorageDevConfig{Path: tempDir}
	serverCfg := config.ServerConfig{Scheme: "http", Host: "localhost", Port: "8080"}
	provider, err := NewLocalStorageProvider(&cfg, &serverCfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Upload first
	_, err = provider.Upload(context.Background(), "exists.txt", strings.NewReader("content"), 7, "text/plain")
	if err != nil {
		t.Fatalf("failed to upload: %v", err)
	}

	exists, err := provider.Exists(context.Background(), "exists.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("expected true, got false")
	}

	exists, err = provider.Exists(context.Background(), "nonexistent.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Errorf("expected false, got true")
	}
}

func TestLocalStorageProvider_GetURL(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.StorageDevConfig{Path: tempDir}
	serverCfg := config.ServerConfig{Scheme: "http", Host: "localhost", Port: "8080"}
	provider, err := NewLocalStorageProvider(&cfg, &serverCfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	url, err := provider.GetURL(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "http://localhost:8080/uploads/test.txt"
	if url != expected {
		t.Errorf("got %v, want %v", url, expected)
	}

	url, err = provider.GetURL(context.Background(), "path\\test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected = "http://localhost:8080/uploads/path/test.txt"
	if url != expected {
		t.Errorf("got %v, want %v", url, expected)
	}
}

func TestLocalStorageProvider_HealthCheck(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.StorageDevConfig{Path: tempDir}
	serverCfg := config.ServerConfig{Scheme: "http", Host: "localhost", Port: "8080"}
	provider, err := NewLocalStorageProvider(&cfg, &serverCfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	err = provider.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test error case by removing the directory
	os.RemoveAll(tempDir)
	err = provider.HealthCheck(context.Background())
	if err == nil {
		t.Fatalf("expected error after removing directory, got nil")
	}
}
