package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Key : key used for initialization and authentication of nodes
type Key struct {
	RaftNodeID uint64 `json:"RAFT_NODE_ID"`
}

// GetKey : gets the node key.json file
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

	var key Key

	if len(data) > 0 {
		keyErr := json.Unmarshal(data, &key)

		if keyErr != nil {
			return nil, keyErr
		}
	}

	return &key, nil
}
