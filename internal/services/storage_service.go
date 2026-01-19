package services

import (
	"context"
	"io"
	"mime/multipart"
	"path/filepath"
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

var reCollapse = regexp.MustCompile(`/+`)

func NormalizePath(path string) string {
	// Convert all backslashes to slashes
	p := strings.ReplaceAll(path, "\\", "/")

	// Collapse multiple slashes into one
	p = reCollapse.ReplaceAllString(p, "/")

	// Clean the path (removes ., .. elements). We intentionally do not
	// allow leading ../ sequences; Clean will reduce them but we also
	// ensure path does not start with '..'
	p = filepath.Clean(p)
	if p == "." {
		return ""
	}
	if strings.HasPrefix(p, "..") {
		return ""
	}
	// Trim any leading / to keep object keys relative
	p = strings.TrimLeft(p, "/")
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
	cleanPath := NormalizePath(path)
	if cleanPath == "" {
		// Use filename only
		cleanPath = file.Filename
	}
	// If original path looks like a directory (ends with slash), append filename
	if strings.HasSuffix(path, "/") {
		cleanPath = filepath.Join(cleanPath, file.Filename)
	}

	url, err := s.provider.Upload(ctx, NormalizePath(cleanPath), src, file.Size, contentType)
	if err != nil {
		return UploadResponse{}, err
	}

	return UploadResponse{
		URL:        url,
		ObjectName: cleanPath,
	}, nil
}

func (s *StorageService) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.provider.Download(ctx, NormalizePath(path))
}

func (s *StorageService) DeleteFile(ctx context.Context, path string) error {
	return s.provider.Delete(ctx, NormalizePath(path))
}

func (s *StorageService) FileExists(ctx context.Context, path string) (bool, error) {
	return s.provider.Exists(ctx, NormalizePath(path))
}

func (s *StorageService) GetFileURL(ctx context.Context, path string) (string, error) {
	return s.provider.GetURL(ctx, NormalizePath(path))
}

func (s *StorageService) HealthCheck(ctx context.Context) error {
	return s.provider.HealthCheck(ctx)
}
