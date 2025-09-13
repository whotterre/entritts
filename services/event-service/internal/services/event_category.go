package services

import (
	"errors"
	"event-service/internal/dto"
	"event-service/internal/models"
	"event-service/internal/repository"
	"strings"

	"github.com/google/uuid"
)

type EventCategoryService interface {
	CreateCategory(request dto.CreateCategoryRequest) (*models.EventCategory, error)
	GetCategoryByID(categoryId uuid.UUID) (*models.EventCategory, error)
	GetAllCategories() ([]models.EventCategory, error)
	UpdateCategory(categoryId uuid.UUID, request dto.UpdateCategoryRequest) (*models.EventCategory, error)
	DeleteCategory(categoryId uuid.UUID) error
}

type eventCategoryService struct {
	categoryRepository repository.EventCategoryRepository
}

func NewEventCategoryService(categoryRepository repository.EventCategoryRepository) EventCategoryService {
	return &eventCategoryService{
		categoryRepository: categoryRepository,
	}
}

func (s *eventCategoryService) CreateCategory(request dto.CreateCategoryRequest) (*models.EventCategory, error) {
	// Validate input
	if strings.TrimSpace(request.Name) == "" {
		return nil, errors.New("category name is required")
	}

	// Check if category with same name already exists
	existingCategory, err := s.categoryRepository.GetCategoryByName(request.Name)
	if err != nil {
		return nil, err
	}
	if existingCategory != nil {
		return nil, errors.New("category with this name already exists")
	}

	// Create new category
	category := &models.EventCategory{
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
	}

	return s.categoryRepository.CreateCategory(category)
}

func (s *eventCategoryService) GetCategoryByID(categoryId uuid.UUID) (*models.EventCategory, error) {
	category, err := s.categoryRepository.GetCategoryByID(categoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}
	return category, nil
}

func (s *eventCategoryService) GetAllCategories() ([]models.EventCategory, error) {
	return s.categoryRepository.GetAllCategories()
}

func (s *eventCategoryService) UpdateCategory(categoryId uuid.UUID, request dto.UpdateCategoryRequest) (*models.EventCategory, error) {
	// Check if category exists
	existingCategory, err := s.categoryRepository.GetCategoryByID(categoryId)
	if err != nil {
		return nil, err
	}
	if existingCategory == nil {
		return nil, errors.New("category not found")
	}

	// Validate input
	if strings.TrimSpace(request.Name) == "" {
		return nil, errors.New("category name is required")
	}

	// Check if another category with same name exists
	categoryWithSameName, err := s.categoryRepository.GetCategoryByName(request.Name)
	if err != nil {
		return nil, err
	}
	if categoryWithSameName != nil && categoryWithSameName.CategoryId != categoryId {
		return nil, errors.New("category with this name already exists")
	}

	// Update category
	updatedCategory := &models.EventCategory{
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
	}

	return s.categoryRepository.UpdateCategory(categoryId, updatedCategory)
}

func (s *eventCategoryService) DeleteCategory(categoryId uuid.UUID) error {
	// Check if category exists
	existingCategory, err := s.categoryRepository.GetCategoryByID(categoryId)
	if err != nil {
		return err
	}
	if existingCategory == nil {
		return errors.New("category not found")
	}

	// TODO: Check if category is being used by any events
	// If yes, prevent deletion or handle gracefully

	return s.categoryRepository.DeleteCategory(categoryId)
}
