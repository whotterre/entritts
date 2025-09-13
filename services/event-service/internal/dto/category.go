package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateCategoryRequest - DTO for creating a new event category
type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// UpdateCategoryRequest - DTO for updating an event category
type UpdateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// CategoryResponse - DTO for category response
type CategoryResponse struct {
	CategoryId  uuid.UUID `json:"category_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoriesListResponse - DTO for listing categories
type CategoriesListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int                `json:"total"`
}
