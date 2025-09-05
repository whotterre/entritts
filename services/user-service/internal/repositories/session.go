package repositories

import (
	"user-service/internal/models"

	"gorm.io/gorm"
)

type SessionsRepository interface {
	CreateNewSession(newSession *models.UserSession) (*models.UserSession, error)
}

type sessionsRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionsRepository {
	return &sessionsRepository{db:db}
}

func (r *sessionsRepository) CreateNewSession(newSession *models.UserSession) (*models.UserSession, error) { 
	// Create a new session
	if err := r.db.Create(newSession).Error; err != nil {
		return &models.UserSession{}, err
	}
	return newSession, nil 
}
