// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package tests

import (
	"os"
	"path/filepath"
	filePkg "github.com/aptlogica/sereni-storage-provider/internal/utils/file"
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
			err := filePkg.CreateDirIfNotExists(tt.path, tt.perm)

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
			result, err := filePkg.SafeJoin(tt.base, tt.elems...)

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

func TestCalculateDirSize(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	subdir := filepath.Join(tempDir, "subdir")
	file3 := filepath.Join(subdir, "file3.txt")

	os.WriteFile(file1, []byte("12345"), 0644)      // 5 bytes
	os.WriteFile(file2, []byte("1234567890"), 0644) // 10 bytes
	os.MkdirAll(subdir, 0755)
	os.WriteFile(file3, []byte("abc"), 0644) // 3 bytes

	tests := []struct {
		name         string
		path         string
		expectedSize int64
		expectError  bool
	}{
		{
			name:         "calculate directory size",
			path:         tempDir,
			expectedSize: 18, // 5 + 10 + 3
			expectError:  false,
		},
		{
			name:         "nonexistent directory",
			path:         filepath.Join(tempDir, "nonexistent"),
			expectedSize: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := filePkg.CalculateDirSize(tt.path)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if size != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, size)
			}
		})
	}
}
