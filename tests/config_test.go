// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package tests

import (
	"os"
	"path/filepath"
	configPkg "github.com/aptlogica/sereni-storage-provider/internal/config"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		envFile  string
		expected configPkg.Config
	}{
		{
			name: "default config",
			expected: configPkg.Config{
				Server: configPkg.ServerConfig{
					Port:               "8080",
					Host:               "localhost",
					IP:                 "localhost",
					Scheme:             "http",
					MaxUploadSizeBytes: 10 << 20, // 10 MiB
				},
				Storage: configPkg.StorageConfig{
					Driver: "local",
					Dev: configPkg.StorageDevConfig{
						Path: "./uploads",
					},
				},
			},
		},
		{
			name: "config with env vars",
			envVars: map[string]string{
				"SERVER_PORT":           "9090",
				"SERVER_HOST":           "example.com",
				"SERVER_IP":             "api.example.com",
				"SERVER_SCHEME":         "https",
				"MAX_UPLOAD_SIZE_BYTES": "20",
				"STORAGE_DRIVER":        "minio",
				"STORAGE_DEV_PATH":      "/tmp/test",
				"MINIO_ENDPOINT":        "minio.example.com",
				"MINIO_ACCESS_KEY":      "key",
				"MINIO_SECRET_KEY":      "secret",
				"MINIO_BUCKET":          "bucket",
				"MINIO_USE_SSL":         "true",
				"AWS_REGION":            "us-east-1",
				"AWS_BUCKET":            "aws-bucket",
				"AWS_ACCESS_KEY":        "aws-key",
				"AWS_SECRET_KEY":        "aws-secret",
				"AWS_ENDPOINT":          "s3.example.com",
			},
			expected: configPkg.Config{
				Server: configPkg.ServerConfig{
					Port:               "9090",
					Host:               "example.com",
					IP:                 "api.example.com",
					Scheme:             "https",
					MaxUploadSizeBytes: 20,
				},
				Storage: configPkg.StorageConfig{
					Driver: "minio",
					Dev: configPkg.StorageDevConfig{
						Path: "/tmp/test",
					},
					Minio: configPkg.StorageMinioConfig{
						Endpoint:  "minio.example.com",
						AccessKey: "key",
						SecretKey: "secret",
						Bucket:    "bucket",
						UseSSL:    true,
					},
					AWS: configPkg.StorageAWSConfig{
						Region:       "us-east-1",
						Bucket:       "aws-bucket",
						AccessKey:    "aws-key",
						SecretKey:    "aws-secret",
						Endpoint:     "s3.example.com",
						UsePathStyle: false,
					},
				},
			},
		},
		{
			name: "config with .env file",
			envFile: `SERVER_PORT=7070
STORAGE_DRIVER=s3
AWS_REGION=us-west-2
`,
			expected: configPkg.Config{
				Server: configPkg.ServerConfig{
					Port:               "7070",
					Host:               "localhost",
					IP:                 "localhost",
					Scheme:             "http",
					MaxUploadSizeBytes: 10 << 20,
				},
				Storage: configPkg.StorageConfig{
					Driver: "s3",
					Dev: configPkg.StorageDevConfig{
						Path: "./uploads",
					},
					AWS: configPkg.StorageAWSConfig{
						Region:       "us-west-2",
						Bucket:       "",
						AccessKey:    "",
						SecretKey:    "",
						Endpoint:     "",
						UsePathStyle: false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalEnv := make(map[string]string)
			for k := range tt.envVars {
				if v, exists := os.LookupEnv(k); exists {
					originalEnv[k] = v
				}
				os.Setenv(k, tt.envVars[k])
			}

			var tempFile string
			if tt.envFile != "" {
				tempFile = filepath.Join(t.TempDir(), ".env")
				os.WriteFile(tempFile, []byte(tt.envFile), 0644)
				originalWd, _ := os.Getwd()
				os.Chdir(filepath.Dir(tempFile))
				defer os.Chdir(originalWd)
			}

			// Act
			cfg, err := configPkg.LoadConfig()

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Server.Port != tt.expected.Server.Port {
				t.Errorf("Server.Port: got %v, want %v", cfg.Server.Port, tt.expected.Server.Port)
			}
			if cfg.Server.Host != tt.expected.Server.Host {
				t.Errorf("Server.Host: got %v, want %v", cfg.Server.Host, tt.expected.Server.Host)
			}
			if cfg.Server.IP != tt.expected.Server.IP {
				t.Errorf("Server.IP: got %v, want %v", cfg.Server.IP, tt.expected.Server.IP)
			}
			if cfg.Server.Scheme != tt.expected.Server.Scheme {
				t.Errorf("Server.Scheme: got %v, want %v", cfg.Server.Scheme, tt.expected.Server.Scheme)
			}
			if cfg.Server.MaxUploadSizeBytes != tt.expected.Server.MaxUploadSizeBytes {
				t.Errorf("Server.MaxUploadSizeBytes: got %v, want %v", cfg.Server.MaxUploadSizeBytes, tt.expected.Server.MaxUploadSizeBytes)
			}
			if cfg.Storage.Driver != tt.expected.Storage.Driver {
				t.Errorf("Storage.Driver: got %v, want %v", cfg.Storage.Driver, tt.expected.Storage.Driver)
			}
			if cfg.Storage.Dev.Path != tt.expected.Storage.Dev.Path {
				t.Errorf("Storage.Dev.Path: got %v, want %v", cfg.Storage.Dev.Path, tt.expected.Storage.Dev.Path)
			}
			if cfg.Storage.Minio.Endpoint != tt.expected.Storage.Minio.Endpoint {
				t.Errorf("Storage.Minio.Endpoint: got %v, want %v", cfg.Storage.Minio.Endpoint, tt.expected.Storage.Minio.Endpoint)
			}
			if cfg.Storage.Minio.AccessKey != tt.expected.Storage.Minio.AccessKey {
				t.Errorf("Storage.Minio.AccessKey: got %v, want %v", cfg.Storage.Minio.AccessKey, tt.expected.Storage.Minio.AccessKey)
			}
			if cfg.Storage.Minio.SecretKey != tt.expected.Storage.Minio.SecretKey {
				t.Errorf("Storage.Minio.SecretKey: got %v, want %v", cfg.Storage.Minio.SecretKey, tt.expected.Storage.Minio.SecretKey)
			}
			if cfg.Storage.Minio.Bucket != tt.expected.Storage.Minio.Bucket {
				t.Errorf("Storage.Minio.Bucket: got %v, want %v", cfg.Storage.Minio.Bucket, tt.expected.Storage.Minio.Bucket)
			}
			if cfg.Storage.Minio.UseSSL != tt.expected.Storage.Minio.UseSSL {
				t.Errorf("Storage.Minio.UseSSL: got %v, want %v", cfg.Storage.Minio.UseSSL, tt.expected.Storage.Minio.UseSSL)
			}
			if cfg.Storage.AWS.Region != tt.expected.Storage.AWS.Region {
				t.Errorf("Storage.AWS.Region: got %v, want %v", cfg.Storage.AWS.Region, tt.expected.Storage.AWS.Region)
			}
			if cfg.Storage.AWS.Bucket != tt.expected.Storage.AWS.Bucket {
				t.Errorf("Storage.AWS.Bucket: got %v, want %v", cfg.Storage.AWS.Bucket, tt.expected.Storage.AWS.Bucket)
			}
			if cfg.Storage.AWS.AccessKey != tt.expected.Storage.AWS.AccessKey {
				t.Errorf("Storage.AWS.AccessKey: got %v, want %v", cfg.Storage.AWS.AccessKey, tt.expected.Storage.AWS.AccessKey)
			}
			if cfg.Storage.AWS.SecretKey != tt.expected.Storage.AWS.SecretKey {
				t.Errorf("Storage.AWS.SecretKey: got %v, want %v", cfg.Storage.AWS.SecretKey, tt.expected.Storage.AWS.SecretKey)
			}
			if cfg.Storage.AWS.Endpoint != tt.expected.Storage.AWS.Endpoint {
				t.Errorf("Storage.AWS.Endpoint: got %v, want %v", cfg.Storage.AWS.Endpoint, tt.expected.Storage.AWS.Endpoint)
			}

			// Cleanup
			for k := range tt.envVars {
				if original, exists := originalEnv[k]; exists {
					os.Setenv(k, original)
				} else {
					os.Unsetenv(k)
				}
			}
		})
	}
}
