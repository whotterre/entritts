package models

import (
	"time"

	"github.com/google/uuid"
)

// UserSessions - keeps info on user sessions
type UserSession struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"session_id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	AccessToken  string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	RefreshToken string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
}
