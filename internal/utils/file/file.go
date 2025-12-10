package file

import (
	"os"
)

func CreateDirIfNotExists(path string, perm os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, perm)
	}
	return nil
}
