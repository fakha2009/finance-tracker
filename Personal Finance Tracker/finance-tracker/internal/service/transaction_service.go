package service

import (
	"errors"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/repository"
	"time"
)

type TransactionService interface {
	CreateTransaction(transaction *models.Transaction) error
	GetUserTransactions(userID int) ([]models.Transaction, error)
	GetUserTransactionsByPeriod(userID int, start, end time.Time) ([]models.Transaction, error)
	GetTransactionByID(id int) (*models.Transaction, error)
	GetTransactionSummary(userID int, start, end time.Time) (*models.TransactionSummary, error)
	GetTransactionsByCategory(userID int, start, end time.Time) ([]models.CategorySummary, error)
	GetMonthlySummary(userID int, year int) ([]models.MonthlySummary, error)
	GetAccountTransactions(accountID int) ([]models.Transaction, error)
}

type transactionService struct {
	repo           repository.Repository
	accountService AccountService
}

func NewTransactionService(repo repository.Repository) TransactionService {
	accountService := NewAccountService(repo)
	return &transactionService{
		repo:           repo,
		accountService: accountService,
	}
}

func (s *transactionService) CreateTransaction(transaction *models.Transaction) error {
	// Проверяем категорию
	category, err := s.repo.GetCategoryByID(transaction.CategoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return errors.New("category not found")
	}

	if category.Type != transaction.Type {
		return errors.New("transaction type does not match category type")
	}

	// Если account_id не указан, используем дефолтный счет пользователя
	if transaction.AccountID == nil {
		defaultAccount, err := s.accountService.GetDefaultAccount(transaction.UserID)
		if err != nil {
			return err
		}
		if defaultAccount == nil {
			return errors.New("no default account found")
		}
		transaction.AccountID = &defaultAccount.ID
	} else {
		// Проверяем, что счет принадлежит пользователю
		account, err := s.accountService.GetAccountByID(*transaction.AccountID)
		if err != nil {
			return err
		}
		if account == nil {
			return errors.New("account not found")
		}
		if account.UserID != transaction.UserID {
			return errors.New("account does not belong to user")
		}
	}

	// Создаем транзакцию
	err = s.repo.CreateTransaction(transaction)
	if err != nil {
		return err
	}

	// Обновляем баланс счета
	err = s.accountService.UpdateAccountBalance(*transaction.AccountID, transaction.Amount, transaction.Type == "income")
	if err != nil {
		return err
	}

	return nil
}

func (s *transactionService) GetUserTransactions(userID int) ([]models.Transaction, error) {
	return s.repo.GetTransactionsByUserID(userID)
}

func (s *transactionService) GetUserTransactionsByPeriod(userID int, start, end time.Time) ([]models.Transaction, error) {
	return s.repo.GetTransactionsByPeriod(userID, start, end)
}

func (s *transactionService) GetTransactionByID(id int) (*models.Transaction, error) {
	return s.repo.GetTransactionByID(id)
}

func (s *transactionService) GetAccountTransactions(accountID int) ([]models.Transaction, error) {
	return s.repo.GetTransactionsByAccountID(accountID)
}

func (s *transactionService) GetTransactionSummary(userID int, start, end time.Time) (*models.TransactionSummary, error) {
	transactions, err := s.GetUserTransactionsByPeriod(userID, start, end)
	if err != nil {
		return nil, err
	}

	var totalIncome, totalExpense float64

	for _, t := range transactions {
		if t.Type == "income" {
			totalIncome += t.Amount
		} else {
			totalExpense += t.Amount
		}
	}

	return &models.TransactionSummary{
		TotalIncome:      totalIncome,
		TotalExpense:     totalExpense,
		NetAmount:        totalIncome - totalExpense,
		TransactionCount: len(transactions),
		PeriodStart:      start.Format("2006-01-02"),
		PeriodEnd:        end.Format("2006-01-02"),
	}, nil
}

func (s *transactionService) GetTransactionsByCategory(userID int, start, end time.Time) ([]models.CategorySummary, error) {
	transactions, err := s.GetUserTransactionsByPeriod(userID, start, end)
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[int]*models.CategorySummary)
	var totalIncome, totalExpense float64

	// Предзагрузка категорий для устранения N+1
	categories, err := s.repo.GetCategoriesByUserID(userID)
	if err == nil {
		for i := range categories {
			c := categories[i]
			categoryMap[c.ID] = &models.CategorySummary{
				CategoryID:   c.ID,
				CategoryName: c.Name,
				Type:         c.Type,
				TotalAmount:  0,
				Count:        0,
			}
		}
	}

	for _, t := range transactions {
		if _, exists := categoryMap[t.CategoryID]; !exists {
			// Фоллбек: одна загрузка при отсутствии предзагрузки
			category, err := s.repo.GetCategoryByID(t.CategoryID)
			if err != nil || category == nil {
				continue
			}
			categoryMap[t.CategoryID] = &models.CategorySummary{
				CategoryID:   t.CategoryID,
				CategoryName: category.Name,
				Type:         category.Type,
				TotalAmount:  0,
				Count:        0,
			}
		}

		categoryMap[t.CategoryID].TotalAmount += t.Amount
		categoryMap[t.CategoryID].Count++

		if t.Type == "income" {
			totalIncome += t.Amount
		} else {
			totalExpense += t.Amount
		}
	}

	result := make([]models.CategorySummary, 0, len(categoryMap))
	for _, summary := range categoryMap {
		var total float64
		if summary.Type == "income" {
			total = totalIncome
		} else {
			total = totalExpense
		}

		if total > 0 {
			summary.Percentage = (summary.TotalAmount / total) * 100
		}

		result = append(result, *summary)
	}

	return result, nil
}

func (s *transactionService) GetMonthlySummary(userID int, year int) ([]models.MonthlySummary, error) {
	result := make([]models.MonthlySummary, 0, 12)

	for month := 1; month <= 12; month++ {
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)

		summary, err := s.GetTransactionSummary(userID, start, end)
		if err != nil {
			continue
		}

		monthly := models.MonthlySummary{
			Month:        start.Format("2006-01"),
			TotalIncome:  summary.TotalIncome,
			TotalExpense: summary.TotalExpense,
			NetAmount:    summary.NetAmount,
		}

		result = append(result, monthly)
	}

	return result, nil
}
