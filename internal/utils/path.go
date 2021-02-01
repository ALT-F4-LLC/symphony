package utils

import (
	"os"
	"path/filepath"
)

// GetDirPath : resolves full path for directory
func GetDirPath(dir string) (*string, error) {
	path, err := filepath.Abs(dir)

	if err != nil {
		return nil, err
	}

	_, statErr := os.Stat(path)

	if statErr != nil {
		if err := os.MkdirAll(path, 0750); err != nil {
			return nil, err
		}
	}

	return &path, nil
}
