package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirIfNotExists(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		perm        os.FileMode
		expectError bool
	}{
		{
			name:        "create new dir",
			path:        filepath.Join(tempDir, "newdir"),
			perm:        0755,
			expectError: false,
		},
		{
			name:        "dir already exists",
			path:        tempDir,
			perm:        0755,
			expectError: false,
		},
		{
			name:        "file exists",
			path:        filepath.Join(tempDir, "file.txt"),
			perm:        0755,
			expectError: true,
		},
	}

	// Create a file for the last test
	os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("test"), 0644)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateDirIfNotExists(tt.path, tt.perm)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check dir exists
			info, err := os.Stat(tt.path)
			if err != nil {
				t.Fatalf("stat failed: %v", err)
			}
			if !info.IsDir() {
				t.Errorf("not a directory")
			}
		})
	}
}

func TestSafeJoin(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		elems       []string
		expected    string
		expectError bool
	}{
		{
			name:        "simple",
			base:        "/base",
			elems:       []string{"file.txt"},
			expected:    "/base/file.txt",
			expectError: false,
		},
		{
			name:        "subdir",
			base:        "/base",
			elems:       []string{"subdir", "file.txt"},
			expected:    "/base/subdir/file.txt",
			expectError: false,
		},
		{
			name:        "path traversal",
			base:        "/base",
			elems:       []string{"../outside.txt"},
			expected:    "",
			expectError: true,
		},
		{
			name:        "absolute path",
			base:        "/base",
			elems:       []string{"/absolute/file.txt"},
			expected:    "/base/absolute/file.txt",
			expectError: false,
		},
		{
			name:        "clean path",
			base:        "/base",
			elems:       []string{"./file.txt"},
			expected:    "/base/file.txt",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeJoin(tt.base, tt.elems...)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
