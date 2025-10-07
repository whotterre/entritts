package services

import (
	"context"
	"errors"
	"time"
	"user-service/internal/dto"
	"user-service/internal/models"
	"user-service/internal/pkg/utils"
	"user-service/internal/repositories"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists       = errors.New("user already exists with that email")
	ErrRoleServiceNotAvailable = errors.New("role service is not available")
)

type UserService interface {
	CreateNewUser(input dto.CreateUserDto, logger *zap.Logger) error
	LoginUser(input dto.LoginUserDto, logger *zap.Logger, jwtSecret string) (dto.LoginUserResponse, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	IncrementUserEventCount(ctx context.Context, userID string) error
}

type userService struct {
	userRepository     repositories.UserRepository
	sessionsRepository repositories.SessionsRepository
}

func NewUserService(userRepository repositories.UserRepository, sessionsRepository repositories.SessionsRepository) UserService {
	return &userService{
		userRepository:     userRepository,
		sessionsRepository: sessionsRepository,
	}
}

func (s *userService) CreateNewUser(input dto.CreateUserDto, logger *zap.Logger) error {
	// Ensure user doesn't exist beforehand
	user, err := s.userRepository.GetUserByEmail(input.Email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Something went wrong while checking if this account already exists")
			return err
		}
	} else if user != nil {
		logger.Error("User already exists with that email")
		return ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password")
		return err
	}

	// Create user model from DTO
	newUser := &models.User{
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PhoneNumber:  input.PhoneNumber,
		PasswordHash: string(hashedPassword),
	}

	err = s.userRepository.CreateUser(newUser)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return err
	}

	logger.Info("User created successfully",
		zap.String("userID", newUser.ID.String()),
	)

	return nil
}

func (s *userService) LoginUser(input dto.LoginUserDto, logger *zap.Logger, pasetoSecret string) (dto.LoginUserResponse, error) {
	// Ensure user exists
	user, err := s.userRepository.GetUserByEmail(input.Email)
	if err != nil {
		logger.Error("Something went wrong while checking if this account already exists")
		return dto.LoginUserResponse{}, err
	}

	if user == nil {
		logger.Error("User not found with that email")
		return dto.LoginUserResponse{}, errors.New("invalid credentials")
	}

	// Debug: Log user details
	logger.Info("Found user", zap.String("user_id", user.ID.String()), zap.String("email", user.Email))

	// Ensure password is valid
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		logger.Error("Failed to compare validity of password")
		return dto.LoginUserResponse{}, err
	}

	// Sign PASETO token
	expiryTime := 1 * time.Hour
	accessToken, err := utils.SignPasetoToken(pasetoSecret, user.ID.String(), user.Email, expiryTime)
	if err != nil {
		return dto.LoginUserResponse{}, err
	}

	refreshToken, err := utils.SignPasetoToken(pasetoSecret, user.ID.String(), user.Email, 168*time.Hour)
	if err != nil {
		return dto.LoginUserResponse{}, err
	}
	// Create a new entry in the sessions table
	newSession := models.UserSession{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(expiryTime),
	}
	session, err := s.sessionsRepository.CreateNewSession(&newSession)
	if err != nil {
		logger.Error("Failed to create new session")
		return dto.LoginUserResponse{}, err
	}

	response := dto.LoginUserResponse{
		AccessToken:  accessToken,
		SessionID:    session.ID.String(),
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(expiryTime),
	}

	return response, nil
}

// GetUserByID retrieves a user by their ID
func (s *userService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	// Parse UUID
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	user, err := s.userRepository.GetUserById(parsedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found, return nil without error
		}
		return nil, err
	}

	return user, nil
}

// IncrementUserEventCount increments the event count for a user (organizer)
func (s *userService) IncrementUserEventCount(ctx context.Context, userID string) error {
	// Parse UUID
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// For now, we'll just log that we're updating the count
	// In a real implementation, you might have an events_count field in the users table
	// or a separate organizer_stats table

	// Check if user exists first
	user, err := s.userRepository.GetUserById(parsedID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	// TODO: Implement actual count increment logic
	// This could be:
	// 1. Update a field in the users table
	// 2. Insert/update a record in an organizer_stats table
	// 3. Cache the count in Redis, etc.

	// For now, this is a no-op that just validates the user exists
	return nil
}
