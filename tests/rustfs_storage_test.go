package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aptlogica/sereni-storage-provider/internal/config"
	rustfsPkg "github.com/aptlogica/sereni-storage-provider/internal/providers/storage/rustfs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type mockRustFSClient struct {
	deleteObjectFunc func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	getObjectFunc    func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	putObjectFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	headObjectFunc   func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	headBucketFunc   func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	listObjectsFunc  func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

func (m *mockRustFSClient) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return m.deleteObjectFunc(ctx, params, optFns...)
}

func (m *mockRustFSClient) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectFunc(ctx, params, optFns...)
}

func (m *mockRustFSClient) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(ctx, params, optFns...)
}

func (m *mockRustFSClient) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return m.headObjectFunc(ctx, params, optFns...)
}

func (m *mockRustFSClient) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return m.headBucketFunc(ctx, params, optFns...)
}

func (m *mockRustFSClient) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if m.listObjectsFunc == nil {
		return &s3.ListObjectsV2Output{Contents: []types.Object{}}, nil
	}
	return m.listObjectsFunc(ctx, params, optFns...)
}

func TestNewRustFSStorageProvider_BucketNotExist(t *testing.T) {
	_, err := rustfsPkg.NewRustFSStorageProvider(nil, nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestRustFSStorageProvider_Delete(t *testing.T) {
	mock := &mockRustFSClient{
		deleteObjectFunc: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
			return &s3.DeleteObjectOutput{}, nil
		},
	}
	provider := &rustfsPkg.RustFSStorageProvider{Client: mock, Bucket: "test-bucket"}
	if err := provider.Delete(context.Background(), "test.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRustFSStorageProvider_Exists(t *testing.T) {
	tests := []struct {
		name           string
		headObjectFunc func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
		expectedExists bool
		expectError    bool
	}{
		{
			name: "exists",
			headObjectFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return &s3.HeadObjectOutput{}, nil
			},
			expectedExists: true,
		},
		{
			name: "not exists",
			headObjectFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return nil, &smithy.GenericAPIError{Code: "NoSuchKey", Message: "object not found"}
			},
			expectedExists: false,
		},
		{
			name: "other error",
			headObjectFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return nil, errors.New("network error")
			},
			expectedExists: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRustFSClient{headObjectFunc: tt.headObjectFunc}
			provider := &rustfsPkg.RustFSStorageProvider{Client: mock, Bucket: "test-bucket"}
			exists, err := provider.Exists(context.Background(), "test.txt")
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.expectedExists {
				t.Fatalf("expected %v, got %v", tt.expectedExists, exists)
			}
		})
	}
}

func TestRustFSStorageProvider_HealthCheck(t *testing.T) {
	mock := &mockRustFSClient{
		headBucketFunc: func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
			return &s3.HeadBucketOutput{}, nil
		},
	}
	provider := &rustfsPkg.RustFSStorageProvider{Client: mock, Bucket: "test-bucket"}
	if err := provider.HealthCheck(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRustFSStorageProvider_GetURL(t *testing.T) {
	provider := &rustfsPkg.RustFSStorageProvider{BaseURL: "https://rustfs.example.com/test-bucket/"}
	url, err := provider.GetURL(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://rustfs.example.com/test-bucket/test.txt" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestRustFSStorageProvider_Download(t *testing.T) {
	mock := &mockRustFSClient{
		getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return nil, errors.New("download error")
		},
	}
	provider := &rustfsPkg.RustFSStorageProvider{Client: mock, Bucket: "test-bucket"}
	_, err := provider.Download(context.Background(), "test.txt")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRustFSStorageProvider_Upload(t *testing.T) {
	mock := &mockRustFSClient{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}
	provider := &rustfsPkg.RustFSStorageProvider{
		Client:  mock,
		Bucket:  "test-bucket",
		BaseURL: "https://rustfs.example.com/test-bucket/",
	}
	result, err := provider.Upload(context.Background(), "test.txt", strings.NewReader("content"), 7, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "https://rustfs.example.com/test-bucket/test.txt" {
		t.Fatalf("unexpected url: %s", result)
	}
}

func TestRustFSStorageProvider_GetSize(t *testing.T) {
	mock := &mockRustFSClient{
		headObjectFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			if aws.ToString(params.Key) == "uploads/" || aws.ToString(params.Key) == "uploads" {
				return nil, &smithy.GenericAPIError{Code: "NoSuchKey", Message: "not found"}
			}
			size := int64(1024)
			return &s3.HeadObjectOutput{ContentLength: &size}, nil
		},
		listObjectsFunc: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			s1 := int64(100)
			s2 := int64(200)
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{{Size: &s1}, {Size: &s2}},
			}, nil
		},
	}
	provider := &rustfsPkg.RustFSStorageProvider{Client: mock, Bucket: "test-bucket"}

	size, isDir, err := provider.GetSize(context.Background(), "file.txt")
	if err != nil || isDir || size != 1024 {
		t.Fatalf("unexpected file size result: size=%d isDir=%v err=%v", size, isDir, err)
	}

	size, isDir, err = provider.GetSize(context.Background(), "uploads/")
	if err != nil || !isDir || size != 300 {
		t.Fatalf("unexpected dir size result: size=%d isDir=%v err=%v", size, isDir, err)
	}

	size, isDir, err = provider.GetSize(context.Background(), "uploads")
	if err != nil || !isDir || size != 300 {
		t.Fatalf("unexpected dir size (no trailing slash) result: size=%d isDir=%v err=%v", size, isDir, err)
	}
}

func TestStorageConfigHasRustFSFields(t *testing.T) {
	cfg := config.StorageConfig{}
	_ = cfg.RustFS
}
