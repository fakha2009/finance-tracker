package service

import (
	"context"
	"errors"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/repository"
)

type AccountService interface {
	CreateAccount(ctx context.Context, account *models.Account) error
	GetUserAccounts(ctx context.Context, userID int) ([]models.Account, error)
	GetAccountByID(ctx context.Context, id int) (*models.Account, error)
	GetDefaultAccount(ctx context.Context, userID int) (*models.Account, error)
	UpdateAccountBalance(accountID int, amount float64, isIncome bool) error
	SetDefaultAccount(ctx context.Context, userID, accountID int) error
}

type accountService struct {
	repo repository.Repository
}

func NewAccountService(repo repository.Repository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) CreateAccount(ctx context.Context, account *models.Account) error {
	// Проверяем существование валюты
	currency, err := s.repo.GetCurrencyByID(ctx, account.CurrencyID)
	if err != nil {
		return err
	}
	if currency == nil {
		return errors.New("currency not found")
	}

	// Если это первый счет пользователя, устанавливаем его как дефолтный
	existingAccounts, err := s.repo.GetAccountsByUserID(ctx, account.UserID)
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

func (s *accountService) GetUserAccounts(ctx context.Context, userID int) ([]models.Account, error) {
	return s.repo.GetAccountsByUserID(ctx, userID)
}

func (s *accountService) GetAccountByID(ctx context.Context, id int) (*models.Account, error) {
	return s.repo.GetAccountByID(ctx, id)
}

func (s *accountService) GetDefaultAccount(ctx context.Context, userID int) (*models.Account, error) {
	return s.repo.GetDefaultAccount(ctx, userID)
}

func (s *accountService) SetDefaultAccount(ctx context.Context, userID, accountID int) error {
	// Проверяем, что счет принадлежит пользователю
	account, err := s.repo.GetAccountByID(ctx, accountID)
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
	account, err := s.repo.GetAccountByID(context.Background(), accountID)
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
