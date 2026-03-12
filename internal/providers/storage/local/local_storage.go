// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
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

func NewLocalStorageProvider(cfg *config.StorageDevConfig, serverCfg *config.ServerConfig) (interfaces.StorageProvider, error) {
	if cfg.Path == "" {
		return nil, errors.New("local storage path is required")
	}

	// Ensure base storage path exists and is absolute
	if err := file.CreateDirIfNotExists(cfg.Path, 0755); err != nil {
		return nil, app_errors.FolderCreateFailed
	}
	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, err
	}

	// Construct base URL for serving files using SERVER_IP
	baseURL := fmt.Sprintf("%s://%s:%s/%s/",
		serverCfg.Scheme,
		serverCfg.IP,
		serverCfg.Port,
		cfg.Path)

	return &LocalStorageProvider{
		path:    absPath,
		baseURL: baseURL,
	}, nil
}

func (l *LocalStorageProvider) Delete(ctx context.Context, objectName string) error {
	// Prevent path traversal and ensure file is within base path
	fullPath, err := file.SafeJoin(l.path, objectName)
	if err != nil {
		return app_errors.ErrInvalidPath
	}

	// Check if the path exists and is a file
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return app_errors.FileNotFound
		}
		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot delete directory: %s is a directory, not a file", objectName)
	}

	if err := os.Remove(fullPath); err != nil {
		return app_errors.ErrDeleteFailed
	}
	return nil
}

func (l *LocalStorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	fullPath, err := file.SafeJoin(l.path, objectName)
	if err != nil {
		return nil, app_errors.ErrInvalidPath
	}
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
	fullPath, err := file.SafeJoin(l.path, objectName)
	if err != nil {
		return false, app_errors.ErrInvalidPath
	}
	_, err = os.Stat(fullPath)
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

	fullPath, err := file.SafeJoin(l.path, objectName)
	if err != nil {
		return "", app_errors.ErrInvalidPath
	}

	// Ensure subdirectory exists
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
	// The static route is configured to serve /uploads from the storage path
	// So we need to prepend /uploads to the object name
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

// GetSize returns the size in bytes of a file or directory
// Returns (size, isDirectory, error)
func (l *LocalStorageProvider) GetSize(ctx context.Context, path string) (int64, bool, error) {
	fullPath, err := file.SafeJoin(l.path, path)
	if err != nil {
		return 0, false, app_errors.ErrInvalidPath
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, app_errors.FileNotFound
		}
		return 0, false, err
	}

	if fileInfo.IsDir() {
		size, err := file.CalculateDirSize(fullPath)
		if err != nil {
			return 0, true, fmt.Errorf("failed to calculate directory size: %w", err)
		}
		return size, true, nil
	}

	return fileInfo.Size(), false, nil
}
