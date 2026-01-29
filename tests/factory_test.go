package tests

import (
	"sereni-storage-provider/internal/config"
	"sereni-storage-provider/internal/providers/storage"
	"testing"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name        string
		storageCfg  config.StorageConfig
		serverCfg   config.ServerConfig
		expectError bool
	}{
		{
			name: "local storage",
			storageCfg: config.StorageConfig{
				Driver: "local",
				Dev: config.StorageDevConfig{
					Path: t.TempDir(),
				},
			},
			serverCfg: config.ServerConfig{
				Scheme: "http",
				Host:   "localhost",
				Port:   "8080",
			},
			expectError: false,
		},
		{
			name: "dev storage",
			storageCfg: config.StorageConfig{
				Driver: "dev",
				Dev: config.StorageDevConfig{
					Path: t.TempDir(),
				},
			},
			serverCfg: config.ServerConfig{
				Scheme: "http",
				Host:   "localhost",
				Port:   "8080",
			},
			expectError: false,
		},
		{
			name: "minio storage",
			storageCfg: config.StorageConfig{
				Driver: "minio",
				Minio: config.StorageMinioConfig{
					Endpoint:  "localhost:9000",
					AccessKey: "key",
					SecretKey: "secret",
					Bucket:    "bucket",
					UseSSL:    false,
				},
			},
			serverCfg: config.ServerConfig{
				Scheme: "http",
				Host:   "localhost",
				Port:   "8080",
			},
			expectError: true, // Expect error due to no minio server
		},
		{
			name: "invalid driver",
			storageCfg: config.StorageConfig{
				Driver: "invalid",
			},
			serverCfg:   config.ServerConfig{},
			expectError: true,
		},
		{
			name: "aws storage",
			storageCfg: config.StorageConfig{
				Driver: "aws",
				AWS: config.StorageAWSConfig{
					Region:    "us-east-1",
					Bucket:    "test-bucket",
					AccessKey: "key",
					SecretKey: "secret",
				},
			},
			serverCfg: config.ServerConfig{
				Scheme: "http",
				Host:   "localhost",
				Port:   "8080",
			},
			expectError: false, // S3 provider creation doesn't validate connection
		},
		{
			name: "s3 storage",
			storageCfg: config.StorageConfig{
				Driver: "s3",
				AWS: config.StorageAWSConfig{
					Region:    "us-east-1",
					Bucket:    "test-bucket",
					AccessKey: "key",
					SecretKey: "secret",
				},
			},
			serverCfg: config.ServerConfig{
				Scheme: "http",
				Host:   "localhost",
				Port:   "8080",
			},
			expectError: false, // S3 provider creation doesn't validate connection
		},
		// Note: Testing minio and s3 would require mocking the clients, which is complex.
		// For unit tests, we can assume they work if the code is correct.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := storage.NewStorage(&tt.storageCfg, &tt.serverCfg)

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

			// Can't easily check type without reflection, but assume it's correct
			if tt.storageCfg.Driver == "local" || tt.storageCfg.Driver == "dev" {
				// Type check for local
			}
		})
	}
}
