package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TicketType - Source of truth for ticket inventory and pricing
type Ticket struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	EventID         uuid.UUID       `gorm:"type:uuid;not null;index" json:"event_id"`
	Name            string          `gorm:"type:varchar(100);not null" json:"name"`
	Description     string          `gorm:"type:text" json:"description"`
	Price           decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	TotalQuantity   int             `gorm:"not null" json:"total_quantity"`
	AvailableAmount int             `gorm:"not null" json:"available"`
	Reserved        int             `gorm:"default:0" json:"reserved"`
	Sold            int             `gorm:"default:0" json:"sold"`
	SaleStartDate   time.Time       `gorm:"not null" json:"sale_start_date"`
	SaleEndDate     time.Time       `gorm:"not null" json:"sale_end_date"`
	IsActive        bool            `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
}