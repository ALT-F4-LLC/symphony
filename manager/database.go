package main

import (
	"github.com/erkrnt/symphony/schema"
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

	if err := tx.FirstOrCreate(&schema.ServiceType{}, schema.ServiceType{Name: "block"}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.FirstOrCreate(&schema.ServiceType{}, schema.ServiceType{Name: "image"}).Error; err != nil {
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
		&schema.LogicalVolume{},
		&schema.PhysicalVolume{},
		&schema.Service{},
		&schema.ServiceType{},
		&schema.VolumeGroup{},
	)

	logrus.Debug("Database auto migrate ran successfully.")

	if flags.Preseed == true {
		preseedDatabase(db)

		logrus.Debug("Database preseed ran successfully.")
	}

	return db, nil
}

func getServiceByHostname(db *gorm.DB, hostname string) (*schema.Service, error) {
	var service schema.Service

	if err := db.Where(&schema.Service{Hostname: hostname}).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &service, nil
}

func getServiceByID(db *gorm.DB, id uuid.UUID) (*schema.Service, error) {
	var service schema.Service

	if err := db.Where("id = ?", id).First(&service).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &service, nil
}

func getServiceTypeByID(db *gorm.DB, id uuid.UUID) (*schema.ServiceType, error) {
	var serviceType schema.ServiceType

	if err := db.Where("id = ?", id).First(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &serviceType, nil
}

func getServiceTypeByName(db *gorm.DB, name string) (*schema.ServiceType, error) {
	serviceType := schema.ServiceType{}

	if err := db.Where(&schema.ServiceType{Name: name}).Find(&serviceType).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &serviceType, nil
}
