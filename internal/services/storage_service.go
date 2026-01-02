package services

import (
	"context"
	"io"
	"mime/multipart"
	"regexp"
	"strings"

	"sereni-storage-provider/internal/providers/storage/interfaces"
)

type StorageService struct {
	provider interfaces.StorageProvider
}

func NewStorageService(provider interfaces.StorageProvider) *StorageService {
	return &StorageService{
		provider: provider,
	}
}

func normalizePath(path string) string {
	// Convert all backslashes to slashes
	p := strings.ReplaceAll(path, "\\", "/")

	// Collapse multiple slashes into one
	re := regexp.MustCompile(`/+`)
	p = re.ReplaceAllString(p, "/")

	return p
}

type UploadResponse struct {
	URL        string
	ObjectName string
}

func (s *StorageService) UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (UploadResponse, error) {
	src, err := file.Open()
	if err != nil {
		return UploadResponse{}, err
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	// If path is not provided, use filename? Or caller provides full path?
	// Let's assume path includes filename or we append it.
	// For now, assuming 'path' is the full object key.

	url, err := s.provider.Upload(ctx, normalizePath(path), src, file.Size, contentType)
	if err != nil {
		return UploadResponse{}, err
	}

	return UploadResponse{
		URL:        url,
		ObjectName: path,
	}, nil
}

func (s *StorageService) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.provider.Download(ctx, normalizePath(path))
}

func (s *StorageService) DeleteFile(ctx context.Context, path string) error {
	return s.provider.Delete(ctx, normalizePath(path))
}

func (s *StorageService) FileExists(ctx context.Context, path string) (bool, error) {
	return s.provider.Exists(ctx, normalizePath(path))
}

func (s *StorageService) GetFileURL(ctx context.Context, path string) (string, error) {
	return s.provider.GetURL(ctx, normalizePath(path))
}

func (s *StorageService) HealthCheck(ctx context.Context) error {
	return s.provider.HealthCheck(ctx)
}
