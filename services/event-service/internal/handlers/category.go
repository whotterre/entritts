package handlers

import (
	"event-service/internal/dto"
	"event-service/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	categoryService services.EventCategoryService
}

func NewCategoryHandler(categoryService services.EventCategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// CreateCategory handles POST /categories
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var request dto.CreateCategoryRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	category, err := h.categoryService.CreateCategory(request)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "category name is required" || err.Error() == "category with this name already exists" {
			status = http.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to create category",
			"details": err.Error(),
		})
	}

	response := dto.CategoryResponse{
		CategoryId:  category.CategoryId,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message":  "Category created successfully",
		"category": response,
	})
}

// GetCategories handles GET /categories
func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve categories",
			"details": err.Error(),
		})
	}

	// Convert to response DTOs
	categoryResponses := make([]dto.CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = dto.CategoryResponse{
			CategoryId:  category.CategoryId,
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   category.CreatedAt,
			UpdatedAt:   category.UpdatedAt,
		}
	}

	response := dto.CategoriesListResponse{
		Categories: categoryResponses,
		Total:      len(categoryResponses),
	}

	return c.Status(http.StatusOK).JSON(response)
}

// GetCategoryByID handles GET /categories/:id
func (h *CategoryHandler) GetCategoryByID(c *fiber.Ctx) error {
	categoryIdStr := c.Params("id")
	categoryId, err := uuid.Parse(categoryIdStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid category ID",
			"details": "Category ID must be a valid UUID",
		})
	}

	category, err := h.categoryService.GetCategoryByID(categoryId)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "category not found" {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to retrieve category",
			"details": err.Error(),
		})
	}

	response := dto.CategoryResponse{
		CategoryId:  category.CategoryId,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	return c.Status(http.StatusOK).JSON(response)
}

// UpdateCategory handles PUT /categories/:id
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	categoryIdStr := c.Params("id")
	categoryId, err := uuid.Parse(categoryIdStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid category ID",
			"details": "Category ID must be a valid UUID",
		})
	}

	var request dto.UpdateCategoryRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	category, err := h.categoryService.UpdateCategory(categoryId, request)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "category not found" {
			status = http.StatusNotFound
		} else if err.Error() == "category name is required" || err.Error() == "category with this name already exists" {
			status = http.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to update category",
			"details": err.Error(),
		})
	}

	response := dto.CategoryResponse{
		CategoryId:  category.CategoryId,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":  "Category updated successfully",
		"category": response,
	})
}

// DeleteCategory handles DELETE /categories/:id
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	categoryIdStr := c.Params("id")
	categoryId, err := uuid.Parse(categoryIdStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid category ID",
			"details": "Category ID must be a valid UUID",
		})
	}

	err = h.categoryService.DeleteCategory(categoryId)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "category not found" {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to delete category",
			"details": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}
