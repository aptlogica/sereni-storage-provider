package tests

import (
	"context"
	"errors"
	"io"
	"net/url"
	"sereni-storage-provider/internal/config"
	minioPkg "sereni-storage-provider/internal/providers/storage/minio"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

// mockMinioError wraps minio.ErrorResponse to implement error interface
type mockMinioError struct {
	minio.ErrorResponse
}

func (e mockMinioError) Error() string {
	return e.Message
}

type mockMinioClient struct {
	removeObjectFunc       func(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error
	getObjectFunc          func(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (*minio.Object, error)
	putObjectFunc          func(ctx context.Context, bucket, object string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	statObjectFunc         func(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	bucketExistsFunc       func(ctx context.Context, bucket string) (bool, error)
	presignedGetObjectFunc func(ctx context.Context, bucket, object string, expiry time.Duration, reqParams url.Values) (*url.URL, error)
	endpointURLFunc        func() *url.URL
}

func (m *mockMinioClient) RemoveObject(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error {
	return m.removeObjectFunc(ctx, bucket, object, opts)
}

func (m *mockMinioClient) GetObject(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return m.getObjectFunc(ctx, bucket, object, opts)
}

func (m *mockMinioClient) PutObject(ctx context.Context, bucket, object string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return m.putObjectFunc(ctx, bucket, object, reader, size, opts)
}

func (m *mockMinioClient) StatObject(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return m.statObjectFunc(ctx, bucket, object, opts)
}

func (m *mockMinioClient) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return m.bucketExistsFunc(ctx, bucket)
}

func (m *mockMinioClient) PresignedGetObject(ctx context.Context, bucket, object string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
	return m.presignedGetObjectFunc(ctx, bucket, object, expiry, reqParams)
}

func (m *mockMinioClient) EndpointURL() *url.URL {
	return m.endpointURLFunc()
}

func TestNewMinioStorageProvider(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.StorageMinioConfig
		expectError bool
	}{
		{
			name: "invalid endpoint",
			cfg: config.StorageMinioConfig{
				Endpoint:  "",
				AccessKey: "key",
				SecretKey: "secret",
				Bucket:    "bucket",
				UseSSL:    false,
			},
			expectError: true,
		},
		{
			name: "valid config but no server",
			cfg: config.StorageMinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				Bucket:    "test-bucket",
				UseSSL:    false,
			},
			expectError: true, // Expect error due to no server
		},
		// Note: For success, it would require a real minio server, so we skip.
		// In a real scenario, use a test minio server or mock the client.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := minioPkg.NewMinioStorageProvider(&tt.cfg)

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

func TestMinioStorageProvider_Delete(t *testing.T) {
	tests := []struct {
		name        string
		object      string
		mockFunc    func(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error
		expectError bool
	}{
		{
			name:   "success",
			object: "test.txt",
			mockFunc: func(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error {
				return nil
			},
			expectError: false,
		},
		{
			name:   "error",
			object: "test.txt",
			mockFunc: func(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error {
				return errors.New("delete error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMinioClient{
				removeObjectFunc: tt.mockFunc,
			}
			provider := &minioPkg.MinioStorageProvider{
				Client: mock,
				Bucket: "test-bucket",
			}
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

func TestMinioStorageProvider_Exists(t *testing.T) {
	tests := []struct {
		name           string
		object         string
		mockFunc       func(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
		expectError    bool
		expectedExists bool
	}{
		{
			name:   "exists",
			object: "test.txt",
			mockFunc: func(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
				return minio.ObjectInfo{}, nil
			},
			expectError:    false,
			expectedExists: true,
		},
		{
			name:   "not exists",
			object: "test.txt",
			mockFunc: func(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
				return minio.ObjectInfo{}, minio.ErrorResponse{Code: "NoSuchKey", Message: "object not found"}
			},
			expectError:    false,
			expectedExists: false,
		},
		{
			name:   "other error",
			object: "test.txt",
			mockFunc: func(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
				return minio.ObjectInfo{}, errors.New("network error")
			},
			expectError:    true,
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMinioClient{
				statObjectFunc: tt.mockFunc,
			}
			provider := &minioPkg.MinioStorageProvider{
				Client: mock,
				Bucket: "test-bucket",
			}
			exists, err := provider.Exists(context.Background(), tt.object)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists != tt.expectedExists {
				t.Errorf("expected exists=%v, got exists=%v", tt.expectedExists, exists)
			}
		})
	}
}

func TestMinioStorageProvider_HealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		mockFunc    func(ctx context.Context, bucket string) (bool, error)
		expectError bool
	}{
		{
			name: "healthy",
			mockFunc: func(ctx context.Context, bucket string) (bool, error) {
				return true, nil
			},
			expectError: false,
		},
		{
			name: "bucket not exists",
			mockFunc: func(ctx context.Context, bucket string) (bool, error) {
				return false, nil
			},
			expectError: true,
		},
		{
			name: "bucket exists error",
			mockFunc: func(ctx context.Context, bucket string) (bool, error) {
				return false, errors.New("network error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMinioClient{
				bucketExistsFunc: tt.mockFunc,
			}
			provider := &minioPkg.MinioStorageProvider{
				Client: mock,
				Bucket: "test-bucket",
			}
			err := provider.HealthCheck(context.Background())
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

func TestMinioStorageProvider_GetURL(t *testing.T) {
	tests := []struct {
		name            string
		endpointURLFunc func() *url.URL
		expectedURL     string
	}{
		{
			name: "public url",
			endpointURLFunc: func() *url.URL {
				return &url.URL{Scheme: "https", Host: "minio.example.com"}
			},
			expectedURL: "https://minio.example.com/test-bucket/test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMinioClient{
				endpointURLFunc: tt.endpointURLFunc,
			}
			provider := &minioPkg.MinioStorageProvider{
				Client: mock,
				Bucket: "test-bucket",
			}
			url, err := provider.GetURL(context.Background(), "test.txt")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if url != tt.expectedURL {
				t.Errorf("expected url %s, got %s", tt.expectedURL, url)
			}
		})
	}
}

func TestMinioStorageProvider_Download(t *testing.T) {
	mock := &mockMinioClient{
		getObjectFunc: func(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (*minio.Object, error) {
			return nil, errors.New("download error")
		},
	}
	provider := &minioPkg.MinioStorageProvider{
		Client: mock,
		Bucket: "test-bucket",
	}
	_, err := provider.Download(context.Background(), "test.txt")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestMinioStorageProvider_Upload(t *testing.T) {
	tests := []struct {
		name          string
		putObjectFunc func(ctx context.Context, bucket, object string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
		expectError   bool
		expectedURL   string
	}{
		{
			name: "upload error",
			putObjectFunc: func(ctx context.Context, bucket, object string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
				return minio.UploadInfo{}, errors.New("upload error")
			},
			expectError: true,
		},
		{
			name: "upload success",
			putObjectFunc: func(ctx context.Context, bucket, object string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
				return minio.UploadInfo{}, nil
			},
			expectError: false,
			expectedURL: "https://minio.example.com/test-bucket/test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMinioClient{
				putObjectFunc:   tt.putObjectFunc,
				endpointURLFunc: func() *url.URL { return &url.URL{Scheme: "https", Host: "minio.example.com"} },
			}
			provider := &minioPkg.MinioStorageProvider{
				Client: mock,
				Bucket: "test-bucket",
			}
			result, err := provider.Upload(context.Background(), "test.txt", strings.NewReader("content"), 7, "text/plain")

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, result)
			}
		})
	}
}
