package main

import (
	"github.com/erkrnt/symphony/services"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Pv : database entry for Pv data
type Pv struct {
	Device string
	ID     uuid.UUID
}

// LoadDB : loads database connection including migrations, etc.
func LoadDB(flags Flags) (*gorm.DB, error) {
	db, err := services.GetDatabase(flags.Debug)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Pv{})
	return db, nil
}
