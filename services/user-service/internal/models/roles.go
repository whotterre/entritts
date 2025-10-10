package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRoles - defines roles for users
type UserRole struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	Role       string    `gorm:"type:varchar(20);not null;check:role IN ('ATTENDEE', 'HOST', 'SPEAKER')" json:"role"`
	AssignedAt time.Time `gorm:"autoCreateTime" json:"assigned_at"`
}
