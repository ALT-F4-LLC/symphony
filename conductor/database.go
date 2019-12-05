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

// Service : struct for service in postgres
type Service struct {
	PrimaryKey
	Hostname      string
	ServiceTypeID uuid.UUID
}

// ServiceType : struct for service in postgres
type ServiceType struct {
	PrimaryKey
	Name string
}

// loadClient : get database connection for service
func loadClient(debug bool) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	if debug == true {
		db.LogMode(true)
	}
	return db, nil
}

// preseed : loads tables with preseed values (types, etc)
func preseed(db *gorm.DB) {
	db.FirstOrCreate(&ServiceType{}, ServiceType{Name: "block"})
}

// BeforeCreate : function for injecting primary keys into DB
func (pk *PrimaryKey) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("ID", uuid.New())
	return nil
}

// GetDatabase : loads database connection including migrations, etc.
func GetDatabase(flags Flags) (*gorm.DB, error) {
	db, err := loadClient(flags.Debug)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Service{}, &ServiceType{})
	if flags.Preseed == true {
		preseed(db)
	}
	return db, nil
}

// GetServiceByHostname : gets specific service from database
func GetServiceByHostname(db *gorm.DB, hostname string) (Services, error) {
	services := make(Services, 0)
	db.Where(&Service{Hostname: hostname}).Find(&services)
	return services, nil
}

// GetServiceTypeByName : gets specific service from database
func GetServiceTypeByName(db *gorm.DB, name string) (ServiceTypes, error) {
	serviceTypes := make(ServiceTypes, 0)
	db.Where(&ServiceType{Name: name}).Find(&serviceTypes)
	return serviceTypes, nil
}
