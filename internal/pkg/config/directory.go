package config

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// GetConfigDir : gets the node root configuration directory
func GetConfigDir(configDir *string) (*string, error) {
	dir := "." // sets default directory

	if *configDir != "" {
		dir = *configDir
	}

	abs, err := filepath.Abs(dir)

	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(abs, 0750); err != nil {
		logrus.Fatal(err)
	}

	return &abs, nil
}
