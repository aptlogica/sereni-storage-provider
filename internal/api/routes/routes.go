// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package routes

import (
	"github.com/aptlogica/sereni-storage-provider/internal/api/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/aptlogica/sereni-storage-provider/docs"
)

func SetupRoutes(router *gin.Engine, storageHandler *handlers.StorageHandler, healthHandler *handlers.HealthHandler, storagePath string) {
	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static Files - serve from configured storage path
	router.Static("/uploads", storagePath)

	api := router.Group("/api/v1")
	{
		api.GET("/health", healthHandler.Health)

		storage := api.Group("/storage")
		{
			storage.POST("/upload", storageHandler.Upload)
			storage.GET("/download", storageHandler.Download)          // /download?path=...
			storage.DELETE("/delete", storageHandler.Delete)           // /delete?path=...
			storage.GET("/consumption", storageHandler.GetConsumption) // /consumption?path=...
		}
	}
}
