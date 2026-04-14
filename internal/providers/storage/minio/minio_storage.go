// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package minio

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aptlogica/sereni-storage-provider/internal/config"
	"github.com/aptlogica/sereni-storage-provider/internal/providers/storage/interfaces"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorageProvider struct {
	Client  interfaces.MinioClient
	Bucket  string
	BaseURL string // Base URL for constructing asset URLs
}

func NewMinioStorageProvider(cfg *config.StorageMinioConfig, serverCfg *config.ServerConfig) (interfaces.StorageProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Verify connection and bucket existence
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check minio bucket existence: %w", err)
	}
	if !exists {
		// Attempt to create bucket if it doesn't exist?
		// For now, let's error out to be safe, or just log.
		// Standard practice: fail if infra isn't ready.
		return nil, fmt.Errorf("minio bucket %s does not exist", cfg.Bucket)
	}

	// Prefer using the MinIO client's endpoint URL for constructing asset base URL
	// so returned URLs point directly to the storage service. Fall back to
	// SERVER_IP + port if the client's endpoint is not available.
	minioPort := client.EndpointURL().Port()
	var baseURL = fmt.Sprintf("%s://%s:%s/%s/", serverCfg.Scheme, serverCfg.IP, minioPort, cfg.Bucket)

	return &MinioStorageProvider{
		Client:  client,
		Bucket:  cfg.Bucket,
		BaseURL: baseURL,
	}, nil
}

func (m *MinioStorageProvider) Delete(ctx context.Context, objectName string) error {
	return m.Client.RemoveObject(ctx, m.Bucket, objectName, minio.RemoveObjectOptions{})
}

func (m *MinioStorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := m.Client.GetObject(ctx, m.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// Check if object is actually readable/exists by reading stat
	_, err = obj.Stat()
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (m *MinioStorageProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.Client.StatObject(ctx, m.Bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MinioStorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// If size is unknown (-1), PutObject might handle it if reader provides Seeker, or we rely on MinIO client multipart.
	// MinIO client requires size. If -1, pass -1 but ensure reader is compatible or use simple streaming if allowed.

	_, err := m.Client.PutObject(ctx, m.Bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return m.GetURL(ctx, objectName)
}

func (m *MinioStorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	// Return the URL using BaseURL (constructed from SERVER_IP) instead of MinIO endpoint
	// This allows the server to proxy/serve MinIO assets through its own address
	cleanPath := strings.ReplaceAll(objectName, "\\", "/")
	return m.BaseURL + cleanPath, nil
}

func (m *MinioStorageProvider) HealthCheck(ctx context.Context) error {
	// Simple connectivity check (bucket exists)
	exists, err := m.Client.BucketExists(ctx, m.Bucket)
	if err != nil {
		return fmt.Errorf("minio health check failed: %w", err)
	}
	if !exists {
		return fmt.Errorf("minio bucket %s does not exist", m.Bucket)
	}
	return nil
}

// GetSize returns the size in bytes of an object or directory in MinIO
// For MinIO, directories are virtual - they are prefixes for object keys
// Returns (size, isDirectory, error)
func (m *MinioStorageProvider) GetSize(ctx context.Context, objectName string) (int64, bool, error) {
	// First try to get as a single object
	stat, err := m.Client.StatObject(ctx, m.Bucket, objectName, minio.StatObjectOptions{})
	if err == nil {
		// Object exists, return its size
		return stat.Size, false, nil
	}

	// If the path ends with "/", treat it as a directory
	if strings.HasSuffix(objectName, "/") {
		size, err := m.getDirectorySize(ctx, objectName)
		return size, true, err
	}

	// Otherwise, return the error
	return 0, false, fmt.Errorf("failed to get object metadata: %w", err)
}

// getDirectorySize calculates the total size of all objects with the given prefix
func (m *MinioStorageProvider) getDirectorySize(ctx context.Context, prefix string) (int64, error) {
	// Ensure prefix ends with "/" for directory-like behavior
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var totalSize int64

	// List all objects with the prefix
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}
	for object := range m.Client.ListObjects(ctx, m.Bucket, opts) {
		if object.Err != nil {
			return 0, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		totalSize += object.Size
	}

	return totalSize, nil
}
