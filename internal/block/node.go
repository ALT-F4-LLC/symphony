package block

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Block : block node
type Block struct {
	Flags *flags
	Key   *key

	mu sync.Mutex
}

// SaveServiceID : sets the SERVICE_ID in key.json
func (b *Block) SaveServiceID(id uuid.UUID) error {
	b.Key.ServiceID = id

	b.mu.Lock()

	keyJSON, err := json.Marshal(b.Key)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/key.json", b.Flags.configDir)

	writeErr := ioutil.WriteFile(path, keyJSON, 0644)

	if writeErr != nil {
		return writeErr
	}

	defer b.mu.Unlock()

	fields := logrus.Fields{
		"SERVICE_ID": id.String(),
	}

	logrus.WithFields(fields).Info("Updated key.json file")

	return nil
}

// New : creates new block node
func New() (*Block, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	key, err := GetKey(flags.configDir)

	if err != nil {
		return nil, err
	}

	b := &Block{
		Flags: flags,
		Key:   key,
	}

	return b, nil
}
