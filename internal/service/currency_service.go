package service

import (
	"context"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/repository"
)

type CurrencyService interface {
	GetAllCurrencies() ([]models.Currency, error)
	GetCurrencyByID(id int) (*models.Currency, error)
	GetCurrencyByCode(code string) (*models.Currency, error)
}

type currencyService struct {
	repo repository.Repository
}

func NewCurrencyService(repo repository.Repository) CurrencyService {
	return &currencyService{repo: repo}
}

func (s *currencyService) GetAllCurrencies() ([]models.Currency, error) {
	return s.repo.GetAllCurrencies()
}

func (s *currencyService) GetCurrencyByID(id int) (*models.Currency, error) {
	return s.repo.GetCurrencyByID(context.Background(), id)
}

func (s *currencyService) GetCurrencyByCode(code string) (*models.Currency, error) {
	return s.repo.GetCurrencyByCode(code)
}
