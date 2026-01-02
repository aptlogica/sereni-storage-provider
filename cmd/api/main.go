package main

import (
	"fmt"
	"log"

	"sereni-storage-provider/internal/api/handlers"
	"sereni-storage-provider/internal/api/routes"
	"sereni-storage-provider/internal/config"
	"sereni-storage-provider/internal/providers/storage"
	"sereni-storage-provider/internal/services"

	"github.com/gin-gonic/gin"
)

// @title Sereni Storage Provider API
// @version 1.0
// @description A robust storage provider service supporting Local, S3, and MinIO backends.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-0.0.html

// @host localhost:8083
// @BasePath /api/v1
func main() {
	// 1. Load Configuration
	config.LoadConfig()

	// 2. Initialize Storage Provider
	provider, err := storage.NewStorage(&config.AppConfig.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize storage provider: %v", err)
	}
	fmt.Printf("Storage Provider Initialized: %s\n", config.AppConfig.Storage.Driver)

	// 3. Initialize Service Layer
	storageService := services.NewStorageService(provider)

	// 4. Initialize Handler Layer
	storageHandler := handlers.NewStorageHandler(storageService)
	healthHandler := handlers.NewHealthHandler(storageService)

	// 5. Setup Router and Routes
	router := gin.Default()

	// Optional: Add CORS, Middleware here
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	routes.SetupRoutes(router, storageHandler, healthHandler, config.AppConfig.Storage.Dev.Path)

	// 6. Start Server
	serverAddr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	fmt.Printf("Starting server on %s...\n", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
