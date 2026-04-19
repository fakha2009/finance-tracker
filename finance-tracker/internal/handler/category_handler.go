package handler

import (
	"net/http"
	"personal-finance-tracker/internal/middleware"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

// DeleteCategory удаляет категорию пользователя
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
    user, exists := middleware.GetUserFromContext(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }

    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category id"})
        return
    }

    if err := h.categoryService.DeleteCategory(id, user.ID); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.Status(http.StatusNoContent)
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories возвращает все категории пользователя
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	categories, err := h.categoryService.GetUserCategories(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// CreateCategory создает новую категорию
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &models.Category{
		UserID:      user.ID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
	}

	if err := h.categoryService.CreateCategory(category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}
