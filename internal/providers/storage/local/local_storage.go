package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	app_errors "sereni-storage-provider/internal/app-errors"
	"sereni-storage-provider/internal/config"
	"sereni-storage-provider/internal/providers/storage/interfaces"
	"sereni-storage-provider/internal/utils/file"
)

type LocalStorageProvider struct {
	path    string
	baseURL string
}

func NewLocalStorageProvider(cfg *config.StorageDevConfig) (interfaces.StorageProvider, error) {
	if cfg.Path == "" {
		return nil, errors.New("local storage path is required")
	}

	err := file.CreateDirIfNotExists(cfg.Path, 0755)
	if err != nil {
		return nil, app_errors.FolderCreateFailed
	}

	// Construct base URL for serving files
	baseURL := fmt.Sprintf("%s://%s:%s/",
		config.AppConfig.Server.Scheme,
		config.AppConfig.Server.Host,
		config.AppConfig.Server.Port)

	return &LocalStorageProvider{
		path:    cfg.Path,
		baseURL: baseURL,
	}, nil
}

func (l *LocalStorageProvider) Delete(ctx context.Context, objectName string) error {
	fullPath := filepath.Join(l.path, objectName)

	// Check if the path exists and is a file
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return app_errors.FileNotFound
		}
		return err
	}

	// Ensure it's a file, not a directory
	if fileInfo.IsDir() {
		return fmt.Errorf("cannot delete directory: %s is a directory, not a file", objectName)
	}

	// Delete the file
	return os.Remove(fullPath)
}

func (l *LocalStorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.path, objectName)
	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, app_errors.FileNotFound
		}
		return nil, err
	}
	return f, nil
}

func (l *LocalStorageProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	fullPath := filepath.Join(l.path, objectName)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LocalStorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// Prevent paths ending with slash (would create directory instead of file)
	if strings.HasSuffix(objectName, "/") || strings.HasSuffix(objectName, "\\") {
		return "", fmt.Errorf("invalid object name: cannot end with path separator")
	}

	fullPath := filepath.Join(l.path, objectName)

	// Ensure subdirectory exists if objectName contains path separators
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", app_errors.FileSaveFailed
	}
	defer f.Close()

	// If size is -1 or unknown, Copy is used, otherwise CopyN
	if size > 0 {
		_, err = io.CopyN(f, reader, size)
	} else {
		_, err = io.Copy(f, reader)
	}

	if err != nil {
		return "", err
	}

	return l.GetURL(ctx, objectName)
}

func (l *LocalStorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	// Simple static file serving URL construction
	// In production this might be different (CDN etc)
	cleanPath := strings.ReplaceAll(objectName, "\\", "/")
	return l.baseURL + cleanPath, nil
}

func (l *LocalStorageProvider) HealthCheck(ctx context.Context) error {
	// Check if storage directory exists and is accessible
	if _, err := os.Stat(l.path); err != nil {
		return fmt.Errorf("local storage path not accessible: %w", err)
	}
	return nil
}
