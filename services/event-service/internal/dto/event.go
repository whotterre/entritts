package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreateNewEventDto represents the request body for creating a new event
type CreateNewEventDto struct {
	OrganizerId string    `json:"organizer_id" validate:"required,uuid"`
	Title       string    `json:"title" validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"required,min=10"`
	CategoryId  uuid.UUID `json:"category_id" validate:"required"`
	StartDate   time.Time `json:"start_date" validate:"required,gtefield=now"`
	EndDate     time.Time `json:"end_date" validate:"required,gtefield=StartDate"`
	VenueId     uuid.UUID `json:"venue_id" validate:"required"`

	MaxCapacity *int                 `json:"max_capacity" validate:"omitempty,min=1"`
	IsPrivate   bool                 `json:"is_private"`
	Tags        []string             `json:"tags" validate:"dive,min=2,max=50"`
	Status      string               `json:"status"`
	SocialLinks []EventSocialLinkDto `json:"social_links" validate:"dive"`
	TicketTypes []TicketType         `json:"ticket_types" validate:"required,dive"`
}

// EventSocialLinkDto represents a social media link for an event
type EventSocialLinkDto struct {
	Platform string `json:"platform" validate:"required,oneof=facebook twitter instagram website linkedin"`
	URL      string `json:"url" validate:"required,url"`
}

// CreateNewEventResponse represents the response after creating a new event
type CreateNewEventResponse struct {
	EventId     uuid.UUID    `json:"event_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	StartDate   time.Time    `json:"start_date"`
	EndDate     time.Time    `json:"end_date"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	TicketTypes []TicketType `json:"ticket_types"`

	// Include basic venue info in response
	Venue struct {
		VenueId   uuid.UUID `json:"venue_id"`
		VenueName string    `json:"venue_name"`
		City      string    `json:"city"`
		Country   string    `json:"country"`
	} `json:"venue"`
}

type TicketType struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	EventID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"event_id"`
	Name          string          `gorm:"type:varchar(100);not null" json:"name"`
	Description   string          `gorm:"type:text" json:"description"`
	Price         decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	TotalQuantity int             `gorm:"not null" json:"total_quantity"`
	Available     int             `gorm:"not null" json:"available"`
	Reserved      int             `gorm:"default:0" json:"reserved"`
	Sold          int             `gorm:"default:0" json:"sold"`
	SaleStartDate time.Time       `gorm:"not null" json:"sale_start_date"`
	SaleEndDate   time.Time       `gorm:"not null" json:"sale_end_date"`
	IsActive      bool            `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateEventStatusDto represents the request to update an event's status
type UpdateEventStatusDto struct {
	Status string `json:"status" validate:"required,oneof=DRAFT PUBLISHED CANCELLED COMPLETED"`
}

// UpdateEventDto represents the request body for updating an existing event
type UpdateEventDto struct {
	Title       *string    `json:"title" validate:"omitempty,min=3,max=255"`
	Description *string    `json:"description" validate:"omitempty,min=10"`
	StartDate   *time.Time `json:"start_date" validate:"omitempty,gtefield=now"`
	EndDate     *time.Time `json:"end_date" validate:"omitempty,gtefield=StartDate"`
	VenueId     *uuid.UUID `json:"venue_id" validate:"omitempty"`
	MaxCapacity *int       `json:"max_capacity" validate:"omitempty,min=1"`
	IsPrivate   *bool      `json:"is_private"`
	Tags        []string   `json:"tags" validate:"omitempty,dive,min=2,max=50"`
}

// GetEventResponse represents the complete event details for retrieval
type GetEventResponse struct {
	EventId     uuid.UUID            `json:"event_id"`
	OrganizerId string               `json:"organizer_id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	CategoryId  uuid.UUID            `json:"category_id"`
	StartDate   time.Time            `json:"start_date"`
	EndDate     time.Time            `json:"end_date"`
	Status      string               `json:"status"`
	MaxCapacity int                  `json:"max_capacity"`
	IsPrivate   bool                 `json:"is_private"`
	Tags        []string             `json:"tags"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Venue       VenueDto             `json:"venue"`
	Category    CategoryDto          `json:"category"`
	SocialLinks []EventSocialLinkDto `json:"social_links"`
}

// VenueDto represents venue information
type VenueDto struct {
	VenueId      uuid.UUID `json:"venue_id"`
	VenueName    string    `json:"venue_name"`
	VenueAddress string    `json:"venue_address"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Capacity     int       `json:"capacity"`
}

// CategoryDto represents category information
type CategoryDto struct {
	CategoryId  uuid.UUID `json:"category_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}
