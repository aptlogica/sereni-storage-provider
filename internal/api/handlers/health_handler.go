package handlers

import (
	"net/http"
	"sereni-storage-provider/internal/services"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	service *services.StorageService
}

func NewHealthHandler(service *services.StorageService) *HealthHandler {
	return &HealthHandler{
		service: service,
	}
}

// Ping godoc
// @Summary Ping check
// @Description Checks if the API is reachable
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "pong"
// @Router /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// Health godoc
// @Summary Health check
// @Description Checks the health of the storage provider
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "OK"
// @Failure 503 {object} map[string]string "Service Unavailable"
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	if err := h.service.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
