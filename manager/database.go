package main

import (
	"github.com/erkrnt/symphony/schemas"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
)

func loadDatabaseClient(verbose bool) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "data.db")

	if err != nil {
		return nil, err
	}

	if verbose == true {
		db.LogMode(true)
	}

	return db, nil
}

func preseedDatabase(db *gorm.DB) error {
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

func getDatabase(flags Flags) (*gorm.DB, error) {
	db, err := loadDatabaseClient(flags.Verbose)

	if err != nil {
		return nil, err
	}

	logrus.Debug("Database client loaded successfully.")

	db.AutoMigrate(
		&schemas.LogicalVolume{},
		&schemas.PhysicalVolume{},
		&schemas.Service{},
		&schemas.ServiceType{},
		&schemas.VolumeGroup{},
	)

	logrus.Debug("Database auto migrate ran successfully.")

	if flags.Preseed == true {
		preseedDatabase(db)

		logrus.Debug("Database preseed ran successfully.")
	}

	return db, nil
}

func getServiceByHostname(db *gorm.DB, hostname string) (*schemas.Service, error) {
	var service schemas.Service

	if err := db.Where(&schemas.Service{Hostname: hostname}).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &service, nil
}

func getServiceByID(db *gorm.DB, id uuid.UUID) (*schemas.Service, error) {
	var service schemas.Service

	if err := db.Where("id = ?", id).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &service, nil
}

func getServiceTypeByID(db *gorm.DB, id uuid.UUID) (*schemas.ServiceType, error) {
	var serviceType schemas.ServiceType

	if err := db.Where("id = ?", id).First(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &serviceType, nil
}

func getServiceTypeByName(db *gorm.DB, name string) (*schemas.ServiceType, error) {
	serviceType := schemas.ServiceType{}

	if err := db.Where(&schemas.ServiceType{Name: name}).Find(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &serviceType, nil
}
