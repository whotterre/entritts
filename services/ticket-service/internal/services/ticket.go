package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"ticket-service/internal/dto"
	"ticket-service/internal/models"
	"ticket-service/internal/repository"

	"go.uber.org/zap"
)

var (
	ErrPermanentInvalidPayload = errors.New("permanent invalid payload")
)

type TicketProducer interface {
	PublishEvent(exchange, routingKey string, payload []byte) error
}

type TicketConsumer interface {
	ConsumeEvents() []byte
}

type TicketService interface {
	CreateNewTicket(input dto.CreateNewTicketDto) (*models.Ticket, error)
	SetConsumer(consumer TicketConsumer)
}

type ticketService struct {
	ticketRepo     repository.TicketRepository
	ticketProducer TicketProducer
	ticketConsumer TicketConsumer
	logger         *zap.Logger
}

func NewTicketService(
	ticketRepo repository.TicketRepository,
	ticketProducer TicketProducer,
	ticketConsumer TicketConsumer,
	logger *zap.Logger) TicketService {
	return &ticketService{
		ticketRepo:     ticketRepo,
		ticketProducer: ticketProducer,
		ticketConsumer: ticketConsumer,
		logger:         logger,
	}
}

func (s *ticketService) SetConsumer(consumer TicketConsumer) {
	s.ticketConsumer = consumer
}

func (s *ticketService) CreateNewTicket(input dto.CreateNewTicketDto) (*models.Ticket, error) {
	// Directly create a ticket record (admin/API flow).
	newTicket := &models.Ticket{
		Name:            input.Name,
		EventID:         input.EventId,
		Description:     input.Description,
		Price:           input.Price,
		TotalQuantity:   input.TotalQuantity,
		AvailableAmount: input.AvailableAmount,
		Reserved:        input.Reserved,
		Sold:            input.Sold,
		SaleStartDate:   input.SaleStartDate,
		SaleEndDate:     input.SaleEndDate,
		IsActive:        input.IsActive,
	}

	createdTicket, err := s.ticketRepo.Create(newTicket)
	if err != nil {
		s.logger.Error("Failed to create new ticket record", zap.Error(err))
		return nil, err
	}
	return createdTicket, nil
}

// ProcessTicketCreateMessage processes a ticket.create message payload and creates a Ticket record.
func (s *ticketService) ProcessTicketCreateMessage(ctx context.Context, payload []byte) error {
	var p map[string]interface{}
	if err := json.Unmarshal(payload, &p); err != nil {
		s.logger.Error("Failed to unmarshal ticket create payload", zap.Error(err))
		return ErrPermanentInvalidPayload
	}

	// Parse required fields
	eventIDStr, ok := p["event_id"].(string)
	if !ok || eventIDStr == "" {
		s.logger.Error("Missing event_id in payload")
		return ErrPermanentInvalidPayload
	}
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		s.logger.Error("Invalid event_id", zap.Error(err))
		return ErrPermanentInvalidPayload
	}

	name, ok := p["name"].(string)
	if !ok || name == "" {
		s.logger.Error("Missing name in payload")
		return ErrPermanentInvalidPayload
	}

	desc, _ := p["description"].(string)

	priceStr, ok := p["price"].(string)
	if !ok || priceStr == "" {
		s.logger.Error("Missing price in payload")
		return ErrPermanentInvalidPayload
	}
	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		s.logger.Error("Invalid price in payload", zap.Error(err))
		return ErrPermanentInvalidPayload
	}

	// total_quantity and available may be floats when unmarshaled from JSON number -> convert
	totalQuantity := 0
	if v, ok := p["total_quantity"].(float64); ok {
		totalQuantity = int(v)
	} else if v, ok := p["total_quantity"].(int); ok {
		totalQuantity = v
	}
	if totalQuantity <= 0 {
		s.logger.Error("Invalid total_quantity in payload")
		return ErrPermanentInvalidPayload
	}

	available := 0
	if v, ok := p["available"].(float64); ok {
		available = int(v)
	}

	// Parse sale dates
	var saleStart, saleEnd time.Time
	if v, ok := p["sale_start_date"].(string); ok && v != "" {
		saleStart, err = time.Parse(time.RFC3339, v)
		if err != nil {
			s.logger.Error("Invalid sale_start_date", zap.Error(err))
			return ErrPermanentInvalidPayload
		}
	}
	if v, ok := p["sale_end_date"].(string); ok && v != "" {
		saleEnd, err = time.Parse(time.RFC3339, v)
		if err != nil {
			s.logger.Error("Invalid sale_end_date", zap.Error(err))
			return ErrPermanentInvalidPayload
		}
	}

	// Transactional create with idempotency check
	tx, err := s.ticketRepo.BeginTx()
	if err != nil {
		s.logger.Error("Failed to begin tx", zap.Error(err))
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// If the payload contains a message_id field, use it for dedupe
	var messageID string
	if mid, ok := p["message_id"].(string); ok {
		messageID = mid
		consumed, err := s.ticketRepo.IsMessageConsumed(tx, messageID)
		if err != nil {
			tx.Rollback()
			s.logger.Error("Consumed message check failed", zap.Error(err))
			return err
		}
		if consumed {
			tx.Commit()
			s.logger.Info("Message already processed, skipping", zap.String("message_id", messageID))
			return nil
		}
	}

	exists, err := s.ticketRepo.ExistsByEventAndName(tx, eventID, name)
	if err != nil {
		tx.Rollback()
		s.logger.Error("Exists check failed", zap.Error(err))
		return err
	}
	if exists {
		tx.Commit()
		s.logger.Info("Ticket already exists, skipping", zap.String("event_id", eventID.String()), zap.String("name", name))
		return nil
	}

	newTicket := &models.Ticket{
		EventID:         eventID,
		Name:            name,
		Description:     desc,
		Price:           price,
		TotalQuantity:   totalQuantity,
		AvailableAmount: available,
		SaleStartDate:   saleStart,
		SaleEndDate:     saleEnd,
		IsActive:        true,
	}

	created, err := s.ticketRepo.CreateWithTx(tx, newTicket)
	if err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create ticket in tx", zap.Error(err))
		return err
	}

	// mark message consumed if messageID present
	if messageID != "" {
		if err := s.ticketRepo.MarkMessageConsumed(tx, messageID); err != nil {
			tx.Rollback()
			s.logger.Error("Failed to mark message consumed", zap.Error(err))
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit ticket tx", zap.Error(err))
		return err
	}

	s.logger.Info("Created ticket from event payload", zap.String("ticket_id", created.ID.String()))
	return nil
}
