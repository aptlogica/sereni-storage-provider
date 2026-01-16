package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"sereni-storage-provider/internal/config"
	"sereni-storage-provider/internal/providers/storage/interfaces"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorageProvider struct {
	Client interfaces.MinioClient
	Bucket string
}

func NewMinioStorageProvider(cfg *config.StorageMinioConfig) (interfaces.StorageProvider, error) {
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

	return &MinioStorageProvider{
		Client: client,
		Bucket: cfg.Bucket,
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
	// Return the public URL for the object
	endpoint := m.Client.EndpointURL()
	return fmt.Sprintf("%s/%s/%s", endpoint.String(), m.Bucket, objectName), nil
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
