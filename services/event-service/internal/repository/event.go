package repository

import (
	"event-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository interface {
	Create(event *models.Event) (*models.Event, error)
	GetEventByID(eventID uuid.UUID) (*models.Event, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *models.Event) (*models.Event, error) {
	if err := r.db.Create(&event).Error; err != nil {
		return nil, err
	}
	return event, nil
}

func (r *eventRepository) GetEventByID(eventID uuid.UUID) (*models.Event, error) {
	var event models.Event
	if err := r.db.Where("event_id = ?", eventID).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}
