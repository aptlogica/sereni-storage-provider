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

func NewStorage(storageCfg *config.StorageConfig, serverCfg *config.ServerConfig) (interfaces.StorageProvider, error) {
	switch strings.ToLower(storageCfg.Driver) {
	case "local", "dev":
		return local.NewLocalStorageProvider(&storageCfg.Dev, serverCfg)

	case "minio":
		return minio.NewMinioStorageProvider(&storageCfg.Minio, serverCfg)

	case "aws", "s3":
		return s3.NewS3StorageProvider(&storageCfg.AWS)

	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", storageCfg.Driver)
	}
}
