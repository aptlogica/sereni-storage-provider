package routes

import (
	"sereni-storage-provider/internal/api/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "sereni-storage-provider/docs"
)

func SetupRoutes(router *gin.Engine, storageHandler *handlers.StorageHandler, healthHandler *handlers.HealthHandler) {
	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static Files
	router.Static("/uploads", "./uploads")

	api := router.Group("/api/v1")
	{
		api.GET("/ping", healthHandler.Ping)
		api.GET("/health", healthHandler.Health)

		storage := api.Group("/storage")
		{
			storage.POST("/upload", storageHandler.Upload)
			storage.GET("/download", storageHandler.Download) // /download?path=...
			storage.DELETE("/delete", storageHandler.Delete)  // /delete?path=...
		}
	}
}
