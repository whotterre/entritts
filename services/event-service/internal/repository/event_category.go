package repository

import (
	"event-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventCategoryRepository interface {
	CreateCategory(category *models.EventCategory) (*models.EventCategory, error)
	GetCategoryByID(categoryId uuid.UUID) (*models.EventCategory, error)
	GetAllCategories() ([]models.EventCategory, error)
	UpdateCategory(categoryId uuid.UUID, category *models.EventCategory) (*models.EventCategory, error)
	DeleteCategory(categoryId uuid.UUID) error
	GetCategoryByName(name string) (*models.EventCategory, error)
}

type eventCategoryRepository struct {
	db *gorm.DB
}

func NewEventCategoryRepository(db *gorm.DB) EventCategoryRepository {
	return &eventCategoryRepository{db: db}
}

func (r *eventCategoryRepository) CreateCategory(category *models.EventCategory) (*models.EventCategory, error) {
	if err := r.db.Create(category).Error; err != nil {
		return nil, err
	}
	return category, nil
}

func (r *eventCategoryRepository) GetCategoryByID(categoryId uuid.UUID) (*models.EventCategory, error) {
	var category models.EventCategory
	if err := r.db.Where("category_id = ?", categoryId).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *eventCategoryRepository) GetAllCategories() ([]models.EventCategory, error) {
	var categories []models.EventCategory
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *eventCategoryRepository) UpdateCategory(categoryId uuid.UUID, category *models.EventCategory) (*models.EventCategory, error) {
	if err := r.db.Where("category_id = ?", categoryId).Updates(category).Error; err != nil {
		return nil, err
	}
	return r.GetCategoryByID(categoryId)
}

func (r *eventCategoryRepository) DeleteCategory(categoryId uuid.UUID) error {
	return r.db.Where("category_id = ?", categoryId).Delete(&models.EventCategory{}).Error
}

func (r *eventCategoryRepository) GetCategoryByName(name string) (*models.EventCategory, error) {
	var category models.EventCategory
	if err := r.db.Where("name = ?", name).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}
