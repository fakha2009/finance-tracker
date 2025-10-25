package service

import (
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/repository"
)

type CategoryService interface {
	CreateCategory(category *models.Category) error
	GetUserCategories(userID int) ([]models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)
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
