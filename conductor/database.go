package main

import (
	"github.com/erkrnt/symphony/schemas"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// LoadDatabaseClient : get database connection for service
func LoadDatabaseClient(verbose bool) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	if verbose == true {
		db.LogMode(true)
	}
	return db, nil
}

// PreseedDatabase : loads tables with preseed values (types, etc)
func PreseedDatabase(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.FirstOrCreate(&schemas.ServiceType{}, schemas.ServiceType{Name: "block"}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.FirstOrCreate(&schemas.ServiceType{}, schemas.ServiceType{Name: "image"}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// GetDatabase : loads database connection including migrations, etc.
func GetDatabase(flags Flags) (*gorm.DB, error) {
	db, err := LoadDatabaseClient(flags.Verbose)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&schemas.PhysicalVolume{}, &schemas.Service{}, &schemas.ServiceType{})
	if flags.Preseed == true {
		PreseedDatabase(db)
	}
	return db, nil
}

// GetServiceByHostname : gets specific service from database
func GetServiceByHostname(db *gorm.DB, hostname string) (*schemas.Service, error) {
	var service schemas.Service
	if err := db.Where(&schemas.Service{Hostname: hostname}).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &service, nil
}

// GetServiceByID : gets specific service from database
func GetServiceByID(db *gorm.DB, id uuid.UUID) (*schemas.Service, error) {
	var service schemas.Service
	if err := db.Where("id = ?", id).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &service, nil
}

// GetServiceTypeByID : get specific service type by ID
func GetServiceTypeByID(db *gorm.DB, id uuid.UUID) (*schemas.ServiceType, error) {
	var serviceType schemas.ServiceType
	if err := db.Where("id = ?", id).First(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &serviceType, nil
}

// GetServiceTypeByName : gets specific service from database
func GetServiceTypeByName(db *gorm.DB, name string) (*schemas.ServiceType, error) {
	serviceType := schemas.ServiceType{}
	if err := db.Where(&schemas.ServiceType{Name: name}).Find(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &serviceType, nil
}

// GetPhysicalVolumeByDeviceID : lookup PhysicalVolume in database from id
func GetPhysicalVolumeByDeviceID(db *gorm.DB, deviceID uuid.UUID) (*schemas.PhysicalVolume, error) {
	physicalVolume := schemas.PhysicalVolume{}
	if err := db.Where("id = ?", deviceID).First(&physicalVolume).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &physicalVolume, nil
}

// GetPhysicalVolumeByDeviceAndServiceID : lookup PhysicalVolume in database from id
func GetPhysicalVolumeByDeviceAndServiceID(db *gorm.DB, device string, serviceID uuid.UUID) (*schemas.PhysicalVolume, error) {
	physicalVolume := schemas.PhysicalVolume{}
	if err := db.Where(&schemas.PhysicalVolume{Device: device, ServiceID: serviceID}).First(&physicalVolume).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &physicalVolume, nil
}

// DeletePhysicalVolumeByDeviceID : lookup PhysicalVolume in database from id
func DeletePhysicalVolumeByDeviceID(db *gorm.DB, deviceID uuid.UUID) error {
	physicalVolume := schemas.PhysicalVolume{}
	if err := db.Where("id = ?", deviceID).Unscoped().Delete(&physicalVolume).Error; err != nil {
		return err
	}
	return nil
}
