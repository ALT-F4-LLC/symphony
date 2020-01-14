package manager

import (
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/flags"
)

// Manager : describes global data for managers
type Manager struct {
	ConfigDir *string
	Key       *config.Key
}

// NewManager : creates a new manager struct
func NewManager(f *flags.ManagerFlags) (*Manager, error) {
	configDir, err := config.GetConfigDir(f.ConfigDir)

	if err != nil {
		return nil, err
	}

	key, err := config.GetKey(*configDir)

	if err != nil {
		return nil, err
	}

	return &Manager{
		ConfigDir: configDir,
		Key:       key,
	}, nil
}
