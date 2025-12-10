package handlers

import (
	"io"
	"net/http"
	"path/filepath"

	"sereni-storage-provider/internal/services"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	service *services.StorageService
}

func NewStorageHandler(service *services.StorageService) *StorageHandler {
	return &StorageHandler{
		service: service,
	}
}

// Upload godoc
// @Summary Upload a file
// @Description Uploads a file to the configured storage provider
// @Tags storage
// @Produce json
// @Param file formData file true "File to upload"
// @Param path formData string false "Path to store the file"
// @Success 200 {object} map[string]string "File uploaded successfully"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /storage/upload [post]
func (h *StorageHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Optional: Allow user to specify path/folder
	path := c.DefaultPostForm("path", "uploads")
	objectName := filepath.Join(path, file.Filename)
	// Sanitize path if needed to prevent directory traversal outside allowed areas?
	// Providers usually handle buckets, but local storage needs watching.
	// For now simple join.

	uploadResponse, err := h.service.UploadFile(c.Request.Context(), file, objectName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"url":     uploadResponse.URL,
		"path":    uploadResponse.ObjectName,
	})
}

// Download godoc
// @Summary Download a file
// @Description Downloads a file from the configured storage provider
// @Tags storage
// @Produce octet-stream
// @Param path query string true "Path to the file"
// @Success 200 {file} file "File content"
// @Failure 404 {object} map[string]string "File Not Found"
// @Failure 400 {object} map[string]string "Bad Request"
// @Router /storage/download [get]
func (h *StorageHandler) Download(c *gin.Context) {
	// objectPath passed as query param or part of URL
	// e.g., /files?path=uploads/image.png
	objectPath := c.Query("path")
	if objectPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}

	reader, err := h.service.GetFile(c.Request.Context(), objectPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or error accessing"})
		return
	}
	defer reader.Close()

	// Detect content type or set default
	// ServeContent could be used if we had Seekable, but we have ReadCloser.
	// We can stream it.

	// Get filename from path
	filename := filepath.Base(objectPath)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	// If we knew the true content type we should store/retrieve it.

	// Stream
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		// Log error, but response might be already committed
	}
}

// Delete godoc
// @Summary Delete a file
// @Description Deletes a file from the configured storage provider
// @Tags storage
// @Produce json
// @Param path query string true "Path to the file"
// @Success 200 {object} map[string]string "File deleted successfully"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /storage/delete [delete]
func (h *StorageHandler) Delete(c *gin.Context) {
	objectPath := c.Query("path")
	if objectPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}

	err := h.service.DeleteFile(c.Request.Context(), objectPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
