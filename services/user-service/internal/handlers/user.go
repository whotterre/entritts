package handlers

import (
	"net/http"
	"user-service/internal/dto"
	"user-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService services.UserService
	logger      *zap.Logger
	jwtSecret   string
}

func NewUserHandler(userService services.UserService, logger *zap.Logger, jwtSecret string) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
		jwtSecret:   jwtSecret,
	}
}

func (h *UserHandler) CreateNewUser(c *fiber.Ctx) error {
	var input dto.CreateUserDto
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Error parsing body while trying to sign up", zap.Error(err))
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

	// Log the user object being created
	h.logger.Info("Creating user",
		zap.String("firstName", input.FirstName),
		zap.String("lastName", input.LastName),
		zap.String("email", input.Email),
		zap.String("phoneNumber", input.PhoneNumber),
	)
	if err := h.userService.CreateNewUser(input, h.logger); err != nil {
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

func (h *UserHandler) LoginUser(c *fiber.Ctx) error {
	var input dto.LoginUserDto

	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Error parsing body while trying to login", zap.Error(err))
	}
	h.logger.Info("Received log in request")
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	h.logger.Info("Parsed input",
		zap.String("email", input.Email),
		zap.String("password", input.Password),
	)

	// Call the service
	response, err := h.userService.LoginUser(input, h.logger, h.jwtSecret)
	if err != nil {
		h.logger.Error("An error occurred while logging in user", zap.Error(err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Login failed",
			"details": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "User logged in successfully",
		"user": fiber.Map{
			"email":        input.Email,
			"accessToken":  response.AccessToken,
			"expiresIn":    response.ExpiresIn,
			"refreshToken": response.RefreshToken,
			"sessionID":    response.SessionID,
		},
	})
}

func (h *UserHandler) GetHealthStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "OK",
		"service": "user-service",
	})
}
