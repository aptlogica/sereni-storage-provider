// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	Storage StorageConfig
}

type ServerConfig struct {
	Port   string
	Host   string
	IP     string // IP/Domain for constructing asset URLs
	Scheme string
	// MaxUploadSizeBytes caps max upload size accepted by server in bytes
	MaxUploadSizeBytes int64
	// AllowedOrigins for CORS
	AllowedOrigins []string
}

type StorageConfig struct {
	Driver string // "local", "s3", "minio"
	Dev    StorageDevConfig
	Minio  StorageMinioConfig
	AWS    StorageAWSConfig
}

type StorageDevConfig struct {
	Path string
}

type StorageMinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type StorageAWSConfig struct {
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
	Endpoint     string // Optional, for custom S3 compatible services
}

// LoadConfig reads configuration from environment and optional .env file
// and returns a Config. It does not mutate package globals.
func LoadConfig() (Config, error) {
	var cfg Config

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Try to read config file, but non-fatal if missing
	if err := viper.ReadInConfig(); err == nil {
		// config file found
	}

	// Set defaults
	viper.SetDefault("STORAGE_DRIVER", "local")
	viper.SetDefault("STORAGE_DEV_PATH", "./uploads")
	viper.SetDefault("SERVER_PORT", "5050")
	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_IP", "localhost")
	viper.SetDefault("SERVER_SCHEME", "http")
	viper.SetDefault("MAX_UPLOAD_SIZE_BYTES", 10<<20) // 10 MiB
	viper.SetDefault("ALLOWED_ORIGINS", "*")

	cfg.Server.Port = viper.GetString("SERVER_PORT")
	cfg.Server.Host = viper.GetString("SERVER_HOST")
	cfg.Server.IP = viper.GetString("SERVER_IP")
	cfg.Server.Scheme = viper.GetString("SERVER_SCHEME")
	cfg.Server.MaxUploadSizeBytes = viper.GetInt64("MAX_UPLOAD_SIZE_BYTES")
	// Parse allowed origins (comma-separated)
	origins := viper.GetString("ALLOWED_ORIGINS")
	if origins == "*" || origins == "" {
		cfg.Server.AllowedOrigins = []string{"*"}
	} else {
		cfg.Server.AllowedOrigins = strings.Split(origins, ",")
		for i := range cfg.Server.AllowedOrigins {
			cfg.Server.AllowedOrigins[i] = strings.TrimSpace(cfg.Server.AllowedOrigins[i])
		}
	}

	cfg.Storage.Driver = viper.GetString("STORAGE_DRIVER")
	cfg.Storage.Dev.Path = viper.GetString("STORAGE_DEV_PATH")

	cfg.Storage.Minio.Endpoint = viper.GetString("MINIO_ENDPOINT")
	cfg.Storage.Minio.AccessKey = viper.GetString("MINIO_ACCESS_KEY")
	cfg.Storage.Minio.SecretKey = viper.GetString("MINIO_SECRET_KEY")
	cfg.Storage.Minio.Bucket = viper.GetString("MINIO_BUCKET")
	cfg.Storage.Minio.UseSSL = viper.GetBool("MINIO_USE_SSL")

	cfg.Storage.AWS.Region = viper.GetString("AWS_REGION")
	cfg.Storage.AWS.Bucket = viper.GetString("AWS_BUCKET")
	cfg.Storage.AWS.AccessKey = viper.GetString("AWS_ACCESS_KEY")
	cfg.Storage.AWS.SecretKey = viper.GetString("AWS_SECRET_KEY")
	cfg.Storage.AWS.Endpoint = viper.GetString("AWS_ENDPOINT")

	return cfg, nil
}
