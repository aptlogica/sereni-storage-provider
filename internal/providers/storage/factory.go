// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package storage

import (
	"fmt"
	"strings"

	"github.com/aptlogica/sereni-storage-provider/internal/config"
	"github.com/aptlogica/sereni-storage-provider/internal/providers/storage/interfaces"
	"github.com/aptlogica/sereni-storage-provider/internal/providers/storage/local"
	"github.com/aptlogica/sereni-storage-provider/internal/providers/storage/rustfs"
	"github.com/aptlogica/sereni-storage-provider/internal/providers/storage/s3"
)

func NewStorage(storageCfg *config.StorageConfig, serverCfg *config.ServerConfig) (interfaces.StorageProvider, error) {
	switch strings.ToLower(storageCfg.Driver) {
	case "local", "dev":
		return local.NewLocalStorageProvider(&storageCfg.Dev, serverCfg)

	case "rustfs":
		return rustfs.NewRustFSStorageProvider(&storageCfg.RustFS, serverCfg)

	case "aws", "s3":
		return s3.NewS3StorageProvider(&storageCfg.AWS)

	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", storageCfg.Driver)
	}
}
