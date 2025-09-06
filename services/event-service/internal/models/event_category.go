package models

import (
	"time"

	"github.com/google/uuid"
)

type EventCategory struct {
	CategoryId  uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"category_id"`
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
}
