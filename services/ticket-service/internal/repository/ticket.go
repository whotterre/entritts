package repository

import (
	"ticket-service/internal/models"

	"github.com/google/uuid"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TicketRepository interface {
	Create(ticket *models.Ticket) (*models.Ticket, error)
	ExistsByEventAndName(tx *gorm.DB, eventID uuid.UUID, name string) (bool, error)
	CreateWithTx(tx *gorm.DB, ticket *models.Ticket) (*models.Ticket, error)
	BeginTx() (*gorm.DB, error)
	// Consumed message helpers
	IsMessageConsumed(tx *gorm.DB, messageID string) (bool, error)
	MarkMessageConsumed(tx *gorm.DB, messageID string) error
}

type ticketRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewTicketRepository(db *gorm.DB, logger *zap.Logger) TicketRepository {
	return &ticketRepository{db: db, logger: logger}
}

// BeginTx starts and returns a new DB transaction
func (r *ticketRepository) BeginTx() (*gorm.DB, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		r.logger.Error("Failed to begin tx", zap.Error(tx.Error))
		return nil, tx.Error
	}
	return tx, nil
}

func (r *ticketRepository) Create(ticket *models.Ticket) (*models.Ticket, error) {
	if err := r.db.Create(&ticket).Error; err != nil {
		r.logger.Error("Failed to create new ticket", zap.Error(err))
		return nil, err
	}
	return ticket, nil
}

// ExistsByEventAndName checks whether a ticket for the given event and name already exists
func (r *ticketRepository) ExistsByEventAndName(tx *gorm.DB, eventID uuid.UUID, name string) (bool, error) {
	var count int64
	if err := tx.Model(&models.Ticket{}).Where("event_id = ? AND name = ?", eventID, name).Count(&count).Error; err != nil {
		r.logger.Error("Failed to check ticket existence", zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// CreateWithTx creates a ticket using the provided transaction
func (r *ticketRepository) CreateWithTx(tx *gorm.DB, ticket *models.Ticket) (*models.Ticket, error) {
	if err := tx.Create(&ticket).Error; err != nil {
		r.logger.Error("Failed to create new ticket (tx)", zap.Error(err))
		return nil, err
	}
	return ticket, nil
}

// IsMessageConsumed checks whether a message ID has already been processed
func (r *ticketRepository) IsMessageConsumed(tx *gorm.DB, messageID string) (bool, error) {
	var count int64
	if err := tx.Model(&models.ConsumedMessage{}).Where("message_id = ?", messageID).Count(&count).Error; err != nil {
		r.logger.Error("Failed to check consumed message", zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// MarkMessageConsumed inserts a consumed_messages row to mark the message as processed
func (r *ticketRepository) MarkMessageConsumed(tx *gorm.DB, messageID string) error {
	cm := &models.ConsumedMessage{
		ID:        uuid.New(),
		MessageID: messageID,
	}
	if err := tx.Create(cm).Error; err != nil {
		r.logger.Error("Failed to mark message consumed", zap.Error(err))
		return err
	}
	return nil
}
