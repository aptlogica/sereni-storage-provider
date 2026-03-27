package local

import (
	"context"
	"os"
	"testing"
)

func TestNewLocalStorageProvider_InvalidPath(t *testing.T) {
	_, err := NewLocalStorageProvider(nil, nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestLocalStorageProvider_Delete_NotExist(t *testing.T) {
	l := &LocalStorageProvider{path: os.TempDir()}
	err := l.Delete(context.Background(), "nonexistent.file")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLocalStorageProvider_Exists_NotExist(t *testing.T) {
	l := &LocalStorageProvider{path: os.TempDir()}
	exists, err := l.Exists(context.Background(), "nonexistent.file")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if exists {
		t.Error("should not exist")
	}
}
