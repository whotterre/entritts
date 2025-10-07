package models

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AggregateID string     `gorm:"type:varchar(255);not null" json:"aggregate_id"`
	EventType   string     `gorm:"type:varchar(255);not null" json:"event_type"`
	EventData   string     `gorm:"type:text;not null" json:"event_data"`
	Published   bool       `gorm:"default:false" json:"published"`
	CreatedAt   time.Time  `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	PublishedAt *time.Time `gorm:"type:timestamp" json:"published_at,omitempty"`
}
