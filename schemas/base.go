package schemas

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// PrimaryKey : primary key schema struct for SQL objects
type PrimaryKey struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;"`
}

// PrimaryTimestamps : primary timestamp struct for SQL objects
type PrimaryTimestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate : function for injecting primary keys into DB
func (pk *PrimaryKey) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("ID", uuid.New())
	return nil
}
