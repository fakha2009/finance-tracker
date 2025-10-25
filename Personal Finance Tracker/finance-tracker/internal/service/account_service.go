package service

import (
	"errors"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/repository"
)

type AccountService interface {
	CreateAccount(account *models.Account) error
	GetUserAccounts(userID int) ([]models.Account, error)
	GetAccountByID(id int) (*models.Account, error)
	GetDefaultAccount(userID int) (*models.Account, error)
	SetDefaultAccount(userID, accountID int) error
	UpdateAccountBalance(accountID int, amount float64, isIncome bool) error
}

type accountService struct {
	repo repository.Repository
}

func NewAccountService(repo repository.Repository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) CreateAccount(account *models.Account) error {
	// Проверяем существование валюты
	currency, err := s.repo.GetCurrencyByID(account.CurrencyID)
	if err != nil {
		return err
	}
	if currency == nil {
		return errors.New("currency not found")
	}

	// Если это первый счет пользователя, устанавливаем его как дефолтный
	existingAccounts, err := s.repo.GetAccountsByUserID(account.UserID)
	if err != nil {
		return err
	}

	if len(existingAccounts) == 0 {
		account.IsDefault = true
	} else if account.IsDefault {
		// Если устанавливаем новый счет как дефолтный, сбрасываем дефолтный статус у других
		err = s.repo.SetDefaultAccount(account.UserID, 0) // 0 означает сброс всех
		if err != nil {
			return err
		}
	}

	return s.repo.CreateAccount(account)
}

func (s *accountService) GetUserAccounts(userID int) ([]models.Account, error) {
	return s.repo.GetAccountsByUserID(userID)
}

func (s *accountService) GetAccountByID(id int) (*models.Account, error) {
	return s.repo.GetAccountByID(id)
}

func (s *accountService) GetDefaultAccount(userID int) (*models.Account, error) {
	return s.repo.GetDefaultAccount(userID)
}

func (s *accountService) SetDefaultAccount(userID, accountID int) error {
	// Проверяем, что счет принадлежит пользователю
	account, err := s.repo.GetAccountByID(accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}
	if account.UserID != userID {
		return errors.New("account does not belong to user")
	}

	return s.repo.SetDefaultAccount(userID, accountID)
}

func (s *accountService) UpdateAccountBalance(accountID int, amount float64, isIncome bool) error {
	account, err := s.repo.GetAccountByID(accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}

	var newBalance float64
	if isIncome {
		newBalance = account.Balance + amount
	} else {
		newBalance = account.Balance - amount
	}

	return s.repo.UpdateAccountBalance(accountID, newBalance)
}
