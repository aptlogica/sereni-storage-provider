package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	Storage StorageConfig
}

type ServerConfig struct {
	Port   string
	Host   string
	Scheme string
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

var AppConfig Config

func LoadConfig() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	} else {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Set defaults
	viper.SetDefault("STORAGE_DRIVER", "local")
	viper.SetDefault("STORAGE_DEV_PATH", "./uploads")
	viper.SetDefault("SERVER_PORT", "5050")
	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_SCHEME", "http")

	// Bind Environment Variables manually if needed or structure them
	// Viper handles nested structs via dot notation in env vars if configured,
	// but standard env vars are usually STORAGE_DRIVER.
	// We will manually map for clarity and robust control.

	AppConfig.Server.Port = viper.GetString("SERVER_PORT")
	AppConfig.Server.Host = viper.GetString("SERVER_HOST")
	AppConfig.Server.Scheme = viper.GetString("SERVER_SCHEME")

	AppConfig.Storage.Driver = viper.GetString("STORAGE_DRIVER")

	// Local
	AppConfig.Storage.Dev.Path = viper.GetString("STORAGE_DEV_PATH")

	// MinIO
	AppConfig.Storage.Minio.Endpoint = viper.GetString("MINIO_ENDPOINT")
	AppConfig.Storage.Minio.AccessKey = viper.GetString("MINIO_ACCESS_KEY")
	AppConfig.Storage.Minio.SecretKey = viper.GetString("MINIO_SECRET_KEY")
	AppConfig.Storage.Minio.Bucket = viper.GetString("MINIO_BUCKET")
	AppConfig.Storage.Minio.UseSSL = viper.GetBool("MINIO_USE_SSL")

	// AWS
	AppConfig.Storage.AWS.Region = viper.GetString("AWS_REGION")
	AppConfig.Storage.AWS.Bucket = viper.GetString("AWS_BUCKET")
	AppConfig.Storage.AWS.AccessKey = viper.GetString("AWS_ACCESS_KEY")
	AppConfig.Storage.AWS.SecretKey = viper.GetString("AWS_SECRET_KEY")
	AppConfig.Storage.AWS.Endpoint = viper.GetString("AWS_ENDPOINT")

	log.Println("Configuration loaded.")
}
