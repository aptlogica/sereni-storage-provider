package handlers

import (
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

	// Enforce server-side size limit
	if h.maxUploadSize > 0 && file.Size > h.maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
		return
	}

	// Optional: Allow user to specify path/folder
	path := c.DefaultPostForm("path", "uploads")

	// Check if path already ends with the filename
	// This handles cases where the client sends the full path including filename
	var objectName string
	if filepath.Base(path) == file.Filename {
		// Path already includes the filename, use as-is
		objectName = path
	} else {
		// Path is just a directory, append filename
		objectName = filepath.Join(path, file.Filename)
	}

	// Validate content-type by reading header bytes from the uploaded file
	fr, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read uploaded file"})
		return
	}
	defer fr.Close()

	// Read first 512 bytes to detect content type
	buf := make([]byte, 512)
	n, _ := fr.Read(buf)
	detected := http.DetectContentType(buf[:n])
	// strip parameters (e.g., ; charset=utf-8)
	if idx := strings.IndexByte(detected, ';'); idx != -1 {
		detected = strings.TrimSpace(detected[:idx])
	}
	if len(h.allowedTypes) > 0 {
		ok := false
		// flexible matching: exact or major-type match (wildcard-like)
		for allowed := range h.allowedTypes {
			if allowed == detected {
				ok = true
				break
			}
			// match major type, e.g., allowed "text/*" or "text/plain" should accept "text/csv"
			if len(allowed) > 0 {
				// split on '/'
				aParts := strings.SplitN(allowed, "/", 2)
				dParts := strings.SplitN(detected, "/", 2)
				if len(aParts) == 2 && len(dParts) == 2 {
					if aParts[1] == "*" && aParts[0] == dParts[0] {
						ok = true
						break
					}
					if aParts[0] == dParts[0] && aParts[1] == dParts[1] {
						ok = true
						break
					}
				}
			}
		}
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content type", "detected": detected})
			return
		}
	}

	uploadResponse, err := h.service.UploadFile(c.Request.Context(), file, objectName)
	if err != nil {
		// Avoid leaking internal errors; return generic message
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
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
		switch err {
		case app_errors.FileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		case app_errors.ErrInvalidPath:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}

	err := h.service.DeleteFile(c.Request.Context(), objectPath)
	if err != nil {
		switch err {
		case app_errors.FileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		case app_errors.ErrInvalidPath:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
