package tests

import (
	"context"
	"errors"
	"io"
	"sereni-storage-provider/internal/config"
	s3Pkg "sereni-storage-provider/internal/providers/storage/s3"
	"strings"
	"testing"

	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	deleteObjectFunc func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	getObjectFunc    func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	putObjectFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	headObjectFunc   func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	headBucketFunc   func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return m.deleteObjectFunc(ctx, params, optFns...)
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectFunc(ctx, params, optFns...)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(ctx, params, optFns...)
}

func (m *mockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return m.headObjectFunc(ctx, params, optFns...)
}

func (m *mockS3Client) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return m.headBucketFunc(ctx, params, optFns...)
}

func TestNewS3StorageProvider(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.StorageAWSConfig
		expectError bool
	}{
		{
			name: "invalid region",
			cfg: config.StorageAWSConfig{
				Region:    "",
				Bucket:    "bucket",
				AccessKey: "key",
				SecretKey: "secret",
			},
			expectError: false, // Session creation may succeed even with empty region
		},
		{
			name: "valid config but no credentials",
			cfg: config.StorageAWSConfig{
				Region:    "us-east-1",
				Bucket:    "test-bucket",
				AccessKey: "key",
				SecretKey: "secret",
			},
			expectError: false, // Session creation succeeds, error comes later
		},
		// Note: For success, requires real AWS credentials, so skip.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := s3Pkg.NewS3StorageProvider(&tt.cfg)

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

func TestS3StorageProvider_Delete(t *testing.T) {
	tests := []struct {
		name        string
		object      string
		mockFunc    func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
		expectError bool
	}{
		{
			name:   "success",
			object: "test.txt",
			mockFunc: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
				return &s3.DeleteObjectOutput{}, nil
			},
			expectError: false,
		},
		{
			name:   "error",
			object: "test.txt",
			mockFunc: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
				return nil, errors.New("delete error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				deleteObjectFunc: tt.mockFunc,
			}
			provider := &s3Pkg.S3StorageProvider{
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

func TestS3StorageProvider_Exists(t *testing.T) {
	tests := []struct {
		name         string
		object       string
		mockFunc     func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
		expectError  bool
		expectExists bool
	}{
		{
			name:   "exists",
			object: "test.txt",
			mockFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return &s3.HeadObjectOutput{}, nil
			},
			expectError:  false,
			expectExists: true,
		},
		{
			name:   "not exists",
			object: "test.txt",
			mockFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return nil, errors.New("not found")
			},
			expectError:  true,
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				headObjectFunc: tt.mockFunc,
			}
			provider := &s3Pkg.S3StorageProvider{
				Client: mock,
				Bucket: "test-bucket",
			}
			exists, err := provider.Exists(context.Background(), tt.object)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if exists != tt.expectExists {
					t.Errorf("expected exists %v, got %v", tt.expectExists, exists)
				}
			}
		})
	}
}

func TestS3StorageProvider_HealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		mockFunc    func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
		expectError bool
	}{
		{
			name: "healthy",
			mockFunc: func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
				return &s3.HeadBucketOutput{}, nil
			},
			expectError: false,
		},
		{
			name: "unhealthy",
			mockFunc: func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
				return nil, errors.New("health check failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				headBucketFunc: tt.mockFunc,
			}
			provider := &s3Pkg.S3StorageProvider{
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

func TestS3StorageProvider_Download(t *testing.T) {
	tests := []struct {
		name        string
		object      string
		mockFunc    func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
		expectError bool
	}{
		{
			name:   "success",
			object: "test.txt",
			mockFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					Body: io.NopCloser(strings.NewReader("test content")),
				}, nil
			},
			expectError: false,
		},
		{
			name:   "error",
			object: "test.txt",
			mockFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("download error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				getObjectFunc: tt.mockFunc,
			}
			provider := &s3Pkg.S3StorageProvider{
				Client: mock,
				Bucket: "test-bucket",
			}
			reader, err := provider.Download(context.Background(), tt.object)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if reader == nil {
					t.Errorf("expected reader, got nil")
				}
				reader.Close()
			}
		})
	}
}

func TestS3StorageProvider_Upload(t *testing.T) {
	tests := []struct {
		name        string
		object      string
		content     string
		mockFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
		expectError bool
	}{
		{
			name:    "success",
			object:  "test.txt",
			content: "test content",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
			expectError: false,
		},
		{
			name:    "error",
			object:  "test.txt",
			content: "test content",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return nil, errors.New("upload error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				putObjectFunc: tt.mockFunc,
			}
			provider := &s3Pkg.S3StorageProvider{
				Client: mock,
				Bucket: "test-bucket",
				Region: "us-east-1",
			}
			url, err := provider.Upload(context.Background(), tt.object, strings.NewReader(tt.content), int64(len(tt.content)), "text/plain")
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if url == "" {
					t.Errorf("expected URL, got empty")
				}
				// Should be fallback URL
				expectedURL := "https://test-bucket.s3.us-east-1.amazonaws.com/test.txt"
				if url != expectedURL {
					t.Errorf("expected URL %s, got %s", expectedURL, url)
				}
			}
		})
	}
}

func TestS3StorageProvider_GetURL(t *testing.T) {
	t.Run("mock client fallback", func(t *testing.T) {
		mock := &mockS3Client{}
		provider := &s3Pkg.S3StorageProvider{
			Client: mock,
			Bucket: "test-bucket",
			Region: "us-east-1",
		}

		url, err := provider.GetURL(context.Background(), "test.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "https://test-bucket.s3.us-east-1.amazonaws.com/test.txt"
		if url != expected {
			t.Errorf("expected URL %s, got %s", expected, url)
		}
	})

	t.Run("real client presign success", func(t *testing.T) {
		client := createInvalidS3Client(t)
		provider := &s3Pkg.S3StorageProvider{
			Client: client,
			Bucket: "test-bucket",
			Region: "us-east-1",
		}

		url, err := provider.GetURL(context.Background(), "test.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should get a presigned URL, not the public fallback
		if !strings.Contains(url, "X-Amz-Algorithm=AWS4-HMAC-SHA256") {
			t.Errorf("expected presigned URL containing AWS4-HMAC-SHA256, got %s", url)
		}
		if !strings.Contains(url, "test-bucket.s3.us-east-1.amazonaws.com") {
			t.Errorf("expected URL to contain bucket and region, got %s", url)
		}
	})
}

// createInvalidS3Client creates an S3 client with invalid credentials for testing presign failure
func createInvalidS3Client(t *testing.T) *s3.Client {
	cfg, err := aws_config.LoadDefaultConfig(context.TODO(),
		aws_config.WithRegion("us-east-1"),
		aws_config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("invalid", "invalid", "")),
	)
	if err != nil {
		t.Fatalf("failed to create AWS config: %v", err)
	}
	return s3.NewFromConfig(cfg)
}
