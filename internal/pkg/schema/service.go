package schema

import (
	"github.com/google/uuid"
)

// Service : struct for service in postgres
type Service struct {
	Hostname      string    `gorm:"unique_index:services_hostname_service_type_id;not null;type:string"`
	ServiceTypeID uuid.UUID `gorm:"unique_index:services_hostname_service_type_id;not null;type:uuid"`
}

// ServiceType : struct for service in postgres
type ServiceType struct {
	Name string `gorm:"unique;not null"`
}
