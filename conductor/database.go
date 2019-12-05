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
	Hostname      string    `gorm:"unique_index:services_hostname_service_type_id;not null;type:string"`
	ServiceTypeID uuid.UUID `gorm:"unique_index:services_hostname_service_type_id;not null;type:uuid"`
}

// ServiceType : struct for service in postgres
type ServiceType struct {
	PrimaryKey
	Name string `gorm:"unique;not null"`
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
func preseed(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.FirstOrCreate(&ServiceType{}, ServiceType{Name: "block"}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
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
func GetServiceByHostname(db *gorm.DB, hostname string) (*Service, error) {
	var service Service
	if err := db.Where(&Service{Hostname: hostname}).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &service, nil
}

// GetServiceByID : gets specific service from database
func GetServiceByID(db *gorm.DB, id string) (*Service, error) {
	var service Service
	if err := db.Where("id = ?", id).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &service, nil
}

// GetServiceTypeByID : get specific service type by ID
func GetServiceTypeByID(db *gorm.DB, id uuid.UUID) (*ServiceType, error) {
	var serviceType ServiceType
	if err := db.Where("id = ?", id).First(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &serviceType, nil
}

// GetServiceTypeByName : gets specific service from database
func GetServiceTypeByName(db *gorm.DB, name string) (*ServiceType, error) {
	serviceType := ServiceType{}
	if err := db.Where(&ServiceType{Name: name}).Find(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &serviceType, nil
}
