package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Key : key used for initialization and authentication of nodes
type Key struct {
	ServiceID *uuid.UUID `json:"service_id"`

	mu sync.Mutex
}

// SaveKeyOptions : used for saving a new key locally
type SaveKeyOptions struct {
	ConfigDir string
	ServiceID uuid.UUID
}

// GetKey : gets contents of key.json file
func GetKey(configDir string) (*Key, error) {
	keyPath := fmt.Sprintf("%s/%s", configDir, "key.json")

	keyFile, err := os.OpenFile(keyPath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	defer keyFile.Close()

	data, err := ioutil.ReadAll(keyFile)

	if err != nil {
		return nil, err
	}

	var key *Key

	if len(data) > 0 {
		keyErr := json.Unmarshal(data, &key)

		if keyErr != nil {
			return nil, keyErr
		}
	}

	if key == nil {
		key = &Key{}
	}

	logrus.Debug("Successfully loaded key.json")

	return key, nil
}

// Save : save the key.json file
func (k *Key) Save(opts SaveKeyOptions) error {
	k.mu.Lock()

	k.ServiceID = &opts.ServiceID

	keyJSON, err := json.Marshal(k)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/key.json", opts.ConfigDir)

	writeErr := ioutil.WriteFile(path, keyJSON, 0644)

	if writeErr != nil {
		return writeErr
	}

	defer k.mu.Unlock()

	fields := logrus.Fields{
		"ServiceID": k.ServiceID.String(),
	}

	logrus.WithFields(fields).Debug("Updated key.json file")

	return nil
}
