package app_errors

import "errors"

var (
	FolderCreateFailed = errors.New("failed to create storage folder")
	FileNotFound       = errors.New("file not found")
	FileSaveFailed     = errors.New("failed to save file")
	ErrInvalidPath     = errors.New("invalid path or path traversal detected")
	ErrDeleteFailed    = errors.New("failed to delete file")
)
