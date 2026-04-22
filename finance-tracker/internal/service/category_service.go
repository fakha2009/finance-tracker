package service

import (
	"handler/internal/models"
	"handler/internal/repository"
)

type CategoryService interface {
	CreateCategory(category *models.Category) error
	GetUserCategories(userID int) ([]models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)
	DeleteCategory(id int, userID int) error
}

type categoryService struct {
	repo repository.Repository
}

func NewCategoryService(repo repository.Repository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(category *models.Category) error {
	return s.repo.CreateCategory(category)
}

func (s *categoryService) GetUserCategories(userID int) ([]models.Category, error) {
	return s.repo.GetCategoriesByUserID(userID)
}

func (s *categoryService) GetCategoryByID(id int) (*models.Category, error) {
	return s.repo.GetCategoryByID(id)
}

func (s *categoryService) DeleteCategory(id int, userID int) error {
	return s.repo.DeleteCategory(id, userID)
}
