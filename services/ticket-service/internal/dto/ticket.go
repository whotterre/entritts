package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateNewTicketDto struct {
	Name          string          `json:"name"`
	EventId       uuid.UUID       `json:"event_id"`
	Description   string          `json:"description"`
	Price         decimal.Decimal `json:"price"`
	TotalQuantity int             `json:"total_quantity"`
	AvailableAmount     int       `json:"available"`
	Reserved      int             `json:"reserved"`
	Sold          int             `json:"sold"`
	SaleStartDate time.Time       `json:"sale_start_date"`
	SaleEndDate   time.Time       `json:"sale_end_date"`
	IsActive      bool            `gorm:"default:true" json:"is_active"`
}