package manager

import (
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/erkrnt/symphony/internal/pkg/flags"
	logger "github.com/erkrnt/symphony/internal/pkg/logs"
	"go.uber.org/zap"
)

// Manager : describes global data for managers
type Manager struct {
	Flags  *flags.ManagerFlags
	Key    *config.Key
	Logger *zap.Logger
	Member *cluster.Member
	Store  *cluster.KvStore
}

// NewManager : creates a new manager struct
func NewManager() (*Manager, error) {
	l := logger.NewLogger()

	defer l.Sync()

	f, err := flags.GetManagerFlags(l)

	if err != nil {
		l.Fatal(err.Error())
	}

	k, err := config.GetKey(f.ConfigDir)

	if err != nil {
		return nil, err
	}

	m := &Manager{
		Flags:  f,
		Key:    k,
		Logger: l,
	}

	return m, nil
}
