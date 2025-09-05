package services

import (
	"errors"
	"time"
	"user-service/internal/dto"
	"user-service/internal/models"
	"user-service/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists = errors.New("User already exists with that email")
)

type UserService interface {
	CreateNewUser(input dto.CreateUserDto, logger *zap.Logger) error
	LoginUser(input dto.LoginUserDto, logger *zap.Logger, jwtSecret string) (dto.LoginUserResponse, error)
}

type userService struct {
	userRepository     repositories.UserRepository
	sessionsRepository repositories.SessionsRepository
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &userService{userRepository: userRepository}
}

func (s *userService) CreateNewUser(input dto.CreateUserDto, logger *zap.Logger) error {
	// Ensure user doesn't exist
	user, err := s.userRepository.GetUserByEmail(input.Email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Something went wrong while checking if this account already exists")
			return err
		}
		// Good - user doesn't exist yet
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

	return nil
}

func (s *userService) LoginUser(input dto.LoginUserDto, logger *zap.Logger, jwtSecret string) (dto.LoginUserResponse, error) {
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

	// Ensure password is valid
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		logger.Error("Failed to compare validity of password")
		return dto.LoginUserResponse{}, err
	}

	// Sign JWT token
	expiryTime := 1 * time.Hour // 1 hr
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(expiryTime).Unix,
	}

	// Access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodES256, accessClaims)
	signedAccessToken, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return dto.LoginUserResponse{}, err
	}
	// Refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(168 * time.Hour).Unix(), // 1 week
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return dto.LoginUserResponse{}, err
	}
	// Create a new entry in the sessions table
	newSession := models.UserSession{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
		ExpiresAt:    time.Now().Add(expiryTime),
	}
	session, err := s.sessionsRepository.CreateNewSession(&newSession)
	if err != nil {
		logger.Error("Failed to create new session")
		return dto.LoginUserResponse{}, err
	}

	response := dto.LoginUserResponse{
		AccessToken:  signedAccessToken,
		SessionID:    session.ID.String(),
		RefreshToken: signedRefreshToken,
		ExpiresIn:    time.Now().Add(expiryTime),
	}

	return response, nil
}
