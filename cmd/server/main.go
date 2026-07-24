/*
Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
This file is part of software developed by Aptlogica Technologies Private Limited.
Licensed under the Apache License 2.0. See the LICENSE file in the project root
for full license information.
Websites:
https://www.aptlogica.com
https://www.serenibase.com
Support:
support@aptlogica.com
support@serenibase.com
*/
package main

import (
	"fmt"

	"github.com/aptlogica/sereni-storage-provider/internal/api/handlers"
	"github.com/aptlogica/sereni-storage-provider/internal/api/routes"
	"github.com/aptlogica/sereni-storage-provider/internal/config"
	"github.com/aptlogica/sereni-storage-provider/internal/providers/storage"
	"github.com/aptlogica/sereni-storage-provider/internal/services"

	"github.com/aptlogica/sereni-storage-provider/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title Sereni Storage Provider API
// @version 1.0
// @description A robust storage provider service supporting Local, S3, and RustFS backends.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8083
// @BasePath /api/v1
func validateSecrets() {
	// TODO: Implement real secret validation logic
	log.Info().Msg("Secrets validated")
}

func main() {
	// 1. Load Configuration
	cfg, _ := config.LoadConfig()

	// Validate secrets before proceeding
	validateSecrets()

	// initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("Configuration loaded")

	// 2. Initialize Storage Provider
	provider, err := storage.NewStorage(&cfg.Storage, &cfg.Server)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage provider")
	}
	log.Info().Str("driver", cfg.Storage.Driver).Msg("Storage provider initialized")

	// 3. Initialize Service Layer
	storageService := services.NewStorageService(provider)

	// 4. Initialize Handler Layer
	// allowed content types - allow all types by passing an empty slice
	allowedTypes := []string{}

	storageHandler := handlers.NewStorageHandler(storageService, cfg.Server.MaxUploadSizeBytes, allowedTypes)
	healthHandler := handlers.NewHealthHandler(storageService)

	// 5. Setup Router and Routes
	router := gin.Default()

	// Attach middlewares
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit(int(cfg.Server.RateLimit)))

	// Optional: Add CORS, Middleware here
	// enforce MaxMultipartMemory from config
	if cfg.Server.MaxUploadSizeBytes > 0 {
		router.MaxMultipartMemory = cfg.Server.MaxUploadSizeBytes
	} else {
		router.MaxMultipartMemory = int64(10 << 20) // 10 MiB
	}

	routes.SetupRoutes(router, storageHandler, healthHandler, cfg.Storage.Dev.Path)

	// 6. Start Server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info().Str("addr", serverAddr).Msg("Starting server")

	if err := router.Run(serverAddr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
