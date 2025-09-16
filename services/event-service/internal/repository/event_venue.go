package repository

import (
	"event-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventVenueRepository interface {
	CreateVenue(venue *models.EventVenue) (*models.EventVenue, error)
	GetVenueByID(venueId uuid.UUID) (*models.EventVenue, error)
	GetAllVenues() ([]models.EventVenue, error)
	UpdateVenue(venueId uuid.UUID, venue *models.EventVenue) (*models.EventVenue, error)
	DeleteVenue(venueId uuid.UUID) error
	GetVenueByName(name string) (*models.EventVenue, error)
}

type eventVenueRepository struct {
	db *gorm.DB
}

func NewEventVenueRepository(db *gorm.DB) EventVenueRepository {
	return &eventVenueRepository{db: db}
}

func (r *eventVenueRepository) CreateVenue(venue *models.EventVenue) (*models.EventVenue, error) {
	if err := r.db.Create(venue).Error; err != nil {
		return nil, err
	}
	return venue, nil
}

func (r *eventVenueRepository) GetVenueByID(venueId uuid.UUID) (*models.EventVenue, error) {
	var venue models.EventVenue
	if err := r.db.Where("venue_id = ?", venueId).First(&venue).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &venue, nil
}

func (r *eventVenueRepository) GetAllVenues() ([]models.EventVenue, error) {
	var categories []models.EventVenue
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *eventVenueRepository) UpdateVenue(venueId uuid.UUID, venue *models.EventVenue) (*models.EventVenue, error) {
	if err := r.db.Where("venue_id = ?", venueId).Updates(venue).Error; err != nil {
		return nil, err
	}
	return r.GetVenueByID(venueId)
}

func (r *eventVenueRepository) DeleteVenue(venueId uuid.UUID) error {
	return r.db.Where("venue_id = ?", venueId).Delete(&models.EventVenue{}).Error
}

func (r *eventVenueRepository) GetVenueByName(name string) (*models.EventVenue, error) {
	var venue models.EventVenue
	if err := r.db.Where("name = ?", name).First(&venue).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &venue, nil
}
