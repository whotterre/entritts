package models

import (
	"time"

	"github.com/google/uuid"	
)

// User model - The main entity
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"user_id"`
	FirstName     string    `gorm:"type:varchar(255);not null" json:"first_name"`
	LastName      string    `gorm:"type:varchar(255);not null" json:"last_name"`
	Email         string    `gorm:"type:varchar(128);not null;uniqueIndex" json:"email"`
	PhoneNumber   string    `gorm:"type:varchar(11);not null;uniqueIndex" json:"phone_number"`
	PasswordHash  string    `gorm:"type:varchar(255);not null" json:"-"`
	ProfilePicURL string    `gorm:"type:varchar(255)" json:"profile_pic_url"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
