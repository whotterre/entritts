package models

import (
	"time"

	"github.com/google/uuid"
)

type EventVenue struct {
	VenueID      uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"venue_id"`
	VenueName    string    `gorm:"type:varchar(255);not null" json:"venue_name"`
	VenueAddress string    `gorm:"type:text;not null" json:"venue_address"`
	City         string    `gorm:"type:varchar(100);not null" json:"city"`
	State        string    `gorm:"type:varchar(100);not null" json:"state"`
	Country      string    `gorm:"type:varchar(100);not null" json:"country"`
	Latitude     float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude    float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Capacity     int       `gorm:"type:integer;not null" json:"capacity"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
}
