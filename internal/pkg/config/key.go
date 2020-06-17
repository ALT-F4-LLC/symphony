package config

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
	ClusterID *uuid.UUID `json:"cluster_id"`
	ServiceID *uuid.UUID `json:"service_id"`

	mu sync.Mutex
}

// Save : save the key.json file
func (k *Key) Save(configDir string) error {
	k.mu.Lock()

	keyJSON, err := json.Marshal(k)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/key.json", configDir)

	writeErr := ioutil.WriteFile(path, keyJSON, 0644)

	if writeErr != nil {
		return writeErr
	}

	defer k.mu.Unlock()

	fields := logrus.Fields{
		"cluster_id": k.ClusterID.String(),
		"service_id": k.ServiceID.String(),
	}

	logrus.WithFields(fields).Debug("Updated key.json file")

	return nil
}

// Get : gets contents of key.json file
func (k *Key) Get(configDir string) (*Key, error) {
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

	var key Key

	if len(data) > 0 {
		keyErr := json.Unmarshal(data, &key)

		if keyErr != nil {
			return nil, keyErr
		}
	}

	return &key, nil
}
