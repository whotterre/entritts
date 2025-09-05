package repositories

import (
	"log"
	"user-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserById(id uuid.UUID) (*models.User, error)
	// GetAll()
	// Delete(id string)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	log.Printf("Creating user: %+v", user)
	result := r.db.Create(user)
	if result.Error != nil {
		log.Printf("Error creating user: %v", result.Error)
		return result.Error
	}
	log.Printf("Created user with ID: %v", user.ID)
	return nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetUserById(id uuid.UUID) (*models.User, error) {
	var user *models.User
	if err := r.db.Where("user_id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) DeleteUser(id uuid.UUID) error {
	var user *models.User
	if err := r.db.Where("user_id = ?", id).Delete(&user).Error; err != nil {
		return err
	}
	return nil
}
