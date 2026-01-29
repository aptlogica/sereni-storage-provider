package file

import (
	"os"
	"path/filepath"
	"strings"
)

func CreateDirIfNotExists(path string, perm os.FileMode) error {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return nil
		}
		return os.ErrExist
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(path, perm)
	}
	return err
}

// SafeJoin joins base and elems into a single path and ensures the result
// is within the provided base directory. It prevents directory traversal
// by rejecting any path that would escape the base.
func SafeJoin(base string, elems ...string) (string, error) {
	joined := filepath.Join(elems...)
	cleaned := filepath.Clean(joined)

	// If joined is absolute, make it relative for the check
	if filepath.IsAbs(cleaned) {
		cleaned = strings.TrimPrefix(cleaned, string(filepath.Separator))
	}

	// Resolve relative path from base
	rel, err := filepath.Rel(filepath.Clean(base), filepath.Join(base, cleaned))
	if err != nil {
		return "", err
	}

	// If the relative path starts with '..' then it escapes base
	if strings.HasPrefix(rel, "..") {
		return "", os.ErrPermission
	}
	return filepath.Join(base, cleaned), nil
}

// CalculateDirSize calculates the total size of a directory recursively
func CalculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
