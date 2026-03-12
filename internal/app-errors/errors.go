// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
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
