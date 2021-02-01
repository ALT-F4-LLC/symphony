package manager

import (
	"bytes"
	"fmt"
	"os"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
)

// DiskImageStore : used for storing images to disk
type DiskImageStore struct {
	Path string
}

// ImageStoreSaveOptions : generic configuration for store save handler
type ImageStoreSaveOptions struct {
	Data        bytes.Buffer
	Description string
	File        string
	Name        string
	Size        int64
}

// ImageStore : base image store implementation
type ImageStore interface {
	Save(options ImageStoreSaveOptions) (*api.Image, error)
}

// NewDiskImageStore : creates a new disk image store
func NewDiskImageStore(path string) (*DiskImageStore, error) {
	p, err := utils.GetDirPath(path)

	if err != nil {
		return nil, err
	}

	store := &DiskImageStore{
		Path: *p,
	}

	return store, nil
}

// Save : save handler for disk image store
func (s *DiskImageStore) Save(options ImageStoreSaveOptions) error {
	id := uuid.New()

	path := fmt.Sprintf("%s/%s", s.Path, id.String())

	file, err := os.Create(path)

	if err != nil {
		return err
	}

	_, writeToErr := options.Data.WriteTo(file)

	if writeToErr != nil {
		return writeToErr
	}

	return nil
}
