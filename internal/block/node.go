package block

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Block : block node
type Block struct {
	Flags *flags
	Key   *config.Key

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

	if flags.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	block := &Block{
		Flags: flags,
	}

	return block, nil
}
