package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventId     uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"event_id"`
	OrganizerId uuid.UUID   `gorm:"type:uuid;not null" json:"organizer_id"`
	Title       string      `gorm:"type:varchar(255);not null" json:"title"`
	Description string      `gorm:"type:text" json:"description"`
	CategoryId  uuid.UUID   `gorm:"type:uuid;not null" json:"category_id"`
	VenueId     *uuid.UUID  `gorm:"type:uuid" json:"venue_id"` 
	StartDate   time.Time   `gorm:"type:timestamp;not null" json:"start_date"`
	EndDate     time.Time   `gorm:"type:timestamp;not null" json:"end_date"`
	Status      EventStatus `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"`
	CreatedAt   time.Time   `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
	// Relationships
	Category EventCategory `gorm:"foreignKey:CategoryId" json:"category"`
	Venue    *EventVenue   `gorm:"foreignKey:VenueId" json:"venue,omitempty"`
}
