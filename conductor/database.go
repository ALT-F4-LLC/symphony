package main

import (
	"github.com/erkrnt/symphony/services"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
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

// BeforeCreate : function for injecting primary keys into DB
func (pk *PrimaryKey) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("ID", uuid.New())
	return nil
}

// LoadDB : loads database connection including migrations, etc.
func LoadDB(flags Flags) (*gorm.DB, error) {
	db, err := services.GetDatabase(flags.Debug)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Service{}, &ServiceType{})
	return db, nil
}

// GetServiceByHostname : gets specific service from database
// func GetServiceByHostname(db *sql.DB, hostname string) (Services, error) {
// 	services := make(Services, 0)
// 	rows, queryErr := db.Query("SELECT hostname, id FROM service WHERE hostname = $1", hostname)
// 	if queryErr != nil {
// 		return nil, queryErr
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var service Service
// 		if e := rows.Scan(&service.Hostname, &service.ID); e != nil {
// 			return nil, e
// 		}
// 		services = append(services, service)
// 	}
// 	return services, nil
// }
