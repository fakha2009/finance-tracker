package repository

import (
	"context"
	"personal-finance-tracker/internal/models"
	"time"
)

type Repository interface {
	// User methods
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	UpdateUser(user *models.User) error
	SetUserDefaultCurrency(userID, currencyID int) error

	// Currency methods
	CreateCurrency(currency *models.Currency) error
	GetAllCurrencies() ([]models.Currency, error)
	GetCurrencyByID(ctx context.Context, id int) (*models.Currency, error)
	GetCurrencyByCode(code string) (*models.Currency, error)

	// Account methods
	CreateAccount(account *models.Account) error
	GetAccountsByUserID(ctx context.Context, userID int) ([]models.Account, error)
	GetAccountByID(ctx context.Context, id int) (*models.Account, error)
	GetDefaultAccount(ctx context.Context, userID int) (*models.Account, error)
	UpdateAccountBalance(accountID int, newBalance float64) error
	SetDefaultAccount(userID, accountID int) error

	// Exchange Rate methods
	CreateOrUpdateExchangeRate(rate *models.ExchangeRate) error
	GetExchangeRate(baseCurrencyID, targetCurrencyID int) (*models.ExchangeRate, error)
	GetAllExchangeRates() ([]models.ExchangeRate, error)
	GetExchangeRatesByBaseCurrency(baseCurrencyID int) ([]models.ExchangeRate, error)

	// Category methods
	CreateCategory(ctx context.Context, category *models.Category) error
	GetCategoriesByUserID(ctx context.Context, userID int) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*models.Category, error)

	// Transaction methods
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	GetTransactionsByUserID(ctx context.Context, userID int) ([]models.Transaction, error)
	GetTransactionByID(ctx context.Context, id int) (*models.Transaction, error)
	GetTransactionsByPeriod(ctx context.Context, userID int, start, end time.Time) ([]models.Transaction, error)
	GetTransactionsByAccountID(ctx context.Context, accountID int) ([]models.Transaction, error)
	GetTransactionsByAccountIDAndPeriod(ctx context.Context, accountID int, start, end time.Time) ([]models.Transaction, error)
	GetTransactionSummaryByAccountID(ctx context.Context, accountID int, start, end time.Time) (*models.TransactionSummary, error)

	// Session methods
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByToken(ctx context.Context, token string) (*models.Session, error)
	DeleteSession(token string) error
}
