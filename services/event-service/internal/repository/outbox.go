package repository

import (
	"event-service/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxRepository interface {
	CreateOutboxEvent(tx *gorm.DB, aggregateID, eventType, eventData string) error
	GetUnpublishedEvents() ([]models.OutboxEvent, error)
	MarkAsPublished(eventID uuid.UUID) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) CreateOutboxEvent(tx *gorm.DB, aggregateID, eventType, eventData string) error {
	outboxEvent := &models.OutboxEvent{
		AggregateID: aggregateID,
		EventType:   eventType,
		EventData:   eventData,
		Published:   false,
	}
	return tx.Create(outboxEvent).Error
}

func (r *outboxRepository) GetUnpublishedEvents() ([]models.OutboxEvent, error) {
	var events []models.OutboxEvent
	err := r.db.Where("published = ?", false).Order("created_at ASC").Find(&events).Error
	return events, err
}

func (r *outboxRepository) MarkAsPublished(eventID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.OutboxEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"published":    true,
			"published_at": &now,
		}).Error
}
