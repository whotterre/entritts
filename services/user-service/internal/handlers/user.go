package handlers

import (
	"user-service/internal/dto"
	"user-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService services.UserService
	logger      *zap.Logger
}

func NewUserHandler(userService services.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:     logger,
	}
}

func (h *UserHandler) CreateNewUser(c *fiber.Ctx) error {
	var input dto.CreateUserDto
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Error parsing body", zap.Error(err))
	}
	h.logger.Info("Received registration request")
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Log the parsed input
	h.logger.Info("Parsed input", 
		zap.String("firstName", input.FirstName),
		zap.String("lastName", input.LastName),
		zap.String("email", input.Email),
		zap.String("phoneNumber", input.PhoneNumber),
	)


	newUser := dto.CreateUserDto{
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PhoneNumber:  input.PhoneNumber,
		PasswordHash: input.PasswordHash,
	}

	// Log the user object being created
	h.logger.Info("Creating user", 
		zap.String("firstName", newUser.FirstName),
		zap.String("lastName", newUser.LastName),
		zap.String("email", newUser.Email),
		zap.String("phoneNumber", newUser.PhoneNumber),
	)
	if err := h.userService.CreateNewUser(newUser, h.logger); err != nil {
		if err == services.ErrUserAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "User already exists with that email",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user": fiber.Map{
			"email":     input.Email,
			"firstName": input.FirstName,
			"lastName":  input.LastName,
		},
	})
}

func (h *UserHandler) GetHealthStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "OK",
		"service": "user-service",
	})
}
