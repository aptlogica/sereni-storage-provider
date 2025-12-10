package storage

import (
	"fmt"
	"strings"

	"sereni-storage-provider/internal/config"
	"sereni-storage-provider/internal/providers/storage/interfaces"
	"sereni-storage-provider/internal/providers/storage/local"
	"sereni-storage-provider/internal/providers/storage/minio"
	"sereni-storage-provider/internal/providers/storage/s3"
)

func NewStorage(cfg *config.StorageConfig) (interfaces.StorageProvider, error) {
	switch strings.ToLower(cfg.Driver) {
	case "local", "dev":
		return local.NewLocalStorageProvider(&cfg.Dev)

	case "minio":
		return minio.NewMinioStorageProvider(&cfg.Minio)

	case "aws", "s3":
		return s3.NewS3StorageProvider(&cfg.AWS)

	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}
