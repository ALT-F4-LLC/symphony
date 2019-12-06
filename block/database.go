package main

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// PrimaryKey : primary key schema struct for SQL objects
type PrimaryKey struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;"`
}

// Pv : database entry for Pv data
type Pv struct {
	PrimaryKey
	Device    string    `gorm:"unique_index:pv_device_service_id;not null;type:string"`
	ServiceID uuid.UUID `gorm:"unique_index:pv_device_service_id;not null;type:uuid"`
}

// loadClient : get database connection for service
func loadClient(verbose bool) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	if verbose == true {
		db.LogMode(true)
	}
	return db, nil
}

// BeforeCreate : function for injecting primary keys into DB
func (pk *PrimaryKey) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("ID", uuid.New())
	return nil
}

// GetDatabase : loads database connection including migrations, etc.
func GetDatabase(flags Flags) (*gorm.DB, error) {
	db, err := loadClient(flags.Verbose)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Pv{})
	return db, nil
}
