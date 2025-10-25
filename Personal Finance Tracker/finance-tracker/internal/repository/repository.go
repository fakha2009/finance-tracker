package repository

import (
	"personal-finance-tracker/internal/models"
	"time"
)

type Repository interface {
	// User methods
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUser(user *models.User) error
	SetUserDefaultCurrency(userID, currencyID int) error

	// Currency methods
	CreateCurrency(currency *models.Currency) error
	GetAllCurrencies() ([]models.Currency, error)
	GetCurrencyByID(id int) (*models.Currency, error)
	GetCurrencyByCode(code string) (*models.Currency, error)

	// Account methods
	CreateAccount(account *models.Account) error
	GetAccountsByUserID(userID int) ([]models.Account, error)
	GetAccountByID(id int) (*models.Account, error)
	GetDefaultAccount(userID int) (*models.Account, error)
	UpdateAccountBalance(accountID int, newBalance float64) error
	SetDefaultAccount(userID, accountID int) error

	// Exchange Rate methods
	CreateOrUpdateExchangeRate(rate *models.ExchangeRate) error
	GetExchangeRate(baseCurrencyID, targetCurrencyID int) (*models.ExchangeRate, error)
	GetAllExchangeRates() ([]models.ExchangeRate, error)
	GetExchangeRatesByBaseCurrency(baseCurrencyID int) ([]models.ExchangeRate, error)

	// Category methods
	CreateCategory(category *models.Category) error
	GetCategoriesByUserID(userID int) ([]models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)

	// Transaction methods
	CreateTransaction(transaction *models.Transaction) error
	GetTransactionsByUserID(userID int) ([]models.Transaction, error)
	GetTransactionByID(id int) (*models.Transaction, error)
	GetTransactionsByPeriod(userID int, start, end time.Time) ([]models.Transaction, error)
	GetTransactionsByAccountID(accountID int) ([]models.Transaction, error)

	// Session methods
	CreateSession(session *models.Session) error
	GetSessionByToken(token string) (*models.Session, error)
	DeleteSession(token string) error
}
