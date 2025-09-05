package handlers

import (
	"log"
	"user-service/internal/models"
	"user-service/internal/services"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateNewUser(c *fiber.Ctx) error {
	var newUser models.User
	if err := c.BodyParser(&newUser); err != nil {
		return err
	}

	log.Print(newUser)
	return nil
}

func (h *UserHandler) GetHealthStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "OK",
		"service": "user-service",
	})
}
