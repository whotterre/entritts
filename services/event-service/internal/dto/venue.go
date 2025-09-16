package dto

// CreateVenueRequest - DTO for creating a new event venue
type CreateVenueRequest struct {
	VenueName    string  `json:"venue_name" validate:"required,min=1,max=255"`
	VenueAddress string  `json:"venue_address" validate:"required"`
	City         string  `json:"city" validate:"required,max=100"`
	State        string  `json:"state" validate:"required,max=100"`
	Country      string  `json:"country" validate:"required,max=100"`
	Capacity     int     `json:"capacity" validate:"required,min=1"`
	Latitude     float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude    float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// UpdateVenueRequest - DTO for updating an event venue
type UpdateVenueRequest struct {
	VenueName    string  `json:"venue_name" validate:"required,min=1,max=255"`
	VenueAddress string  `json:"venue_address" validate:"required"`
	City         string  `json:"city" validate:"required,max=100"`
	State        string  `json:"state" validate:"required,max=100"`
	Country      string  `json:"country" validate:"required,max=100"`
	Capacity     int     `json:"capacity" validate:"required,min=1"`
	Latitude     float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude    float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// VenueResponse - DTO for venue response
type VenueResponse struct {
	VenueID      string  `json:"venue_id"`
	VenueName    string  `json:"venue_name"`
	VenueAddress string  `json:"venue_address"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Country      string  `json:"country"`
	Capacity     int     `json:"capacity"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}
