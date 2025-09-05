package services

import (
	"errors"
	"user-service/internal/models"
	"user-service/internal/repositories"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("User already exists with that email")
)

type UserService interface {

}

type userService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &userService{userRepository: userRepository}
}

func (s *userService) CreateNewUser(input *models.User, logger *zap.Logger) error {
	// Ensure user doesn't exist
	user, err := s.userRepository.GetUserByEmail(input.Email)
	if user != nil {
		logger.Error("User already exists with that email")
		return ErrUserAlreadyExists
	}

	if err != nil {
		logger.Error("Something went wrong while checking if this account already exists")
		return err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password")
		return err
	}

	input.PasswordHash = string(hashedPassword)
	
	err = s.userRepository.CreateUser(input)
	if err != nil {
		logger.Error("Failed to create user")
		return err 
	}

	return nil 
}