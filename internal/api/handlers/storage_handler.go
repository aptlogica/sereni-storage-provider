package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	app_errors "sereni-storage-provider/internal/app-errors"
	"sereni-storage-provider/internal/services"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	service       *services.StorageService
	maxUploadSize int64
	allowedTypes  map[string]struct{}
}

const (
	pathParamRequiredMsg = "Path parameter is required"
	invalidPathMsg       = "invalid path"
)

func NewStorageHandler(service *services.StorageService, maxUploadSize int64, allowed []string) *StorageHandler {
	a := make(map[string]struct{}, len(allowed))
	for _, t := range allowed {
		a[t] = struct{}{}
	}
	return &StorageHandler{
		service:       service,
		maxUploadSize: maxUploadSize,
		allowedTypes:  a,
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

	if h.maxUploadSize > 0 && file.Size > h.maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
		return
	}

	path := c.DefaultPostForm("path", "uploads")
	objectName := resolveObjectName(path, file.Filename)

	fr, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read uploaded file"})
		return
	}
	defer fr.Close()

	detected, err := detectAndValidateContentType(fr, h.allowedTypes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "detected": detected})
		return
	}

	uploadResponse, err := h.service.UploadFile(c.Request.Context(), file, objectName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"url":     uploadResponse.URL,
		"path":    uploadResponse.ObjectName,
	})
}

// resolveObjectName determines the final object name for storage
func resolveObjectName(path, filename string) string {
	if filepath.Base(path) == filename {
		return path
	}
	return filepath.Join(path, filename)
}

// detectAndValidateContentType reads the first 512 bytes and validates against allowed types
func detectAndValidateContentType(fr io.ReadSeeker, allowedTypes map[string]struct{}) (string, error) {
	buf := make([]byte, 512)
	n, _ := fr.Read(buf)
	fr.Seek(0, io.SeekStart) // reset for later use
	detected := http.DetectContentType(buf[:n])
	if idx := strings.IndexByte(detected, ';'); idx != -1 {
		detected = strings.TrimSpace(detected[:idx])
	}
	if len(allowedTypes) == 0 {
		return detected, nil
	}
	if isAllowedContentType(detected, allowedTypes) {
		return detected, nil
	}
	return detected, &contentTypeError{msg: "invalid content type"}
}

type contentTypeError struct {
	msg string
}

func (e *contentTypeError) Error() string {
	return e.msg
}

// isAllowedContentType checks if detected type matches allowed types (exact or major type)
func isAllowedContentType(detected string, allowedTypes map[string]struct{}) bool {
	for allowed := range allowedTypes {
		if allowed == detected {
			return true
		}
		aParts := strings.SplitN(allowed, "/", 2)
		dParts := strings.SplitN(detected, "/", 2)
		if len(aParts) == 2 && len(dParts) == 2 {
			if aParts[1] == "*" && aParts[0] == dParts[0] {
				return true
			}
			if aParts[0] == dParts[0] && aParts[1] == dParts[1] {
				return true
			}
		}
	}
	return false
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
		c.JSON(http.StatusBadRequest, gin.H{"error": pathParamRequiredMsg})
		return
	}

	reader, err := h.service.GetFile(c.Request.Context(), objectPath)
	if err != nil {
		switch err {
		case app_errors.FileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		case app_errors.ErrInvalidPath:
			c.JSON(http.StatusBadRequest, gin.H{"error": invalidPathMsg})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to access file"})
			return
		}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": pathParamRequiredMsg})
		return
	}

	err := h.service.DeleteFile(c.Request.Context(), objectPath)
	if err != nil {
		switch err {
		case app_errors.FileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		case app_errors.ErrInvalidPath:
			c.JSON(http.StatusBadRequest, gin.H{"error": invalidPathMsg})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// GetConsumption godoc
// @Summary Get storage consumption
// @Description Returns the size in bytes of a file or directory
// @Tags storage
// @Produce json
// @Param path query string true "Path to the file or directory"
// @Success 200 {object} map[string]interface{} "Storage consumption details"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "File Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /storage/consumption [get]
func (h *StorageHandler) GetConsumption(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": pathParamRequiredMsg})
		return
	}

	size, isDir, err := h.service.GetSize(c.Request.Context(), path)
	if err != nil {
		switch err {
		case app_errors.FileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file or directory not found"})
			return
		case app_errors.ErrInvalidPath:
			c.JSON(http.StatusBadRequest, gin.H{"error": invalidPathMsg})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get size information"})
			return
		}
	}

	response := gin.H{
		"path":         path,
		"size_bytes":   size,
		"is_directory": isDir,
	}

	// Add human-readable size
	response["size_human"] = formatBytes(size)

	c.JSON(http.StatusOK, response)
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
