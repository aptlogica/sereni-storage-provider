// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package app_errors

import "errors"

var (
	FolderCreateFailed      = errors.New("failed to create storage folder")
	FileNotFound            = errors.New("file not found")
	FileSaveFailed          = errors.New("failed to save file")
	ErrInvalidPath          = errors.New("invalid path or path traversal detected")
	ErrDeleteFailed         = errors.New("failed to delete file")
	ErrGenericStorageFailed = errors.New("storage operation failed")
)
