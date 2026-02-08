package models

import (
	"time"
)

type User struct {
	ID                int       `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	Password          string    `json:"-"`
	DefaultCurrencyID *int      `json:"default_currency_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Currency struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Symbol    string    `json:"symbol"`
	CreatedAt time.Time `json:"created_at"`
}

type Account struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	CurrencyID int       `json:"currency_id"`
	Balance    float64   `json:"balance"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Currency   *Currency `json:"currency,omitempty"`
}

type ExchangeRate struct {
	ID               int       `json:"id"`
	BaseCurrencyID   int       `json:"base_currency_id"`
	TargetCurrencyID int       `json:"target_currency_id"`
	Rate             float64   `json:"rate"`
	LastUpdated      time.Time `json:"last_updated"`
	BaseCurrency     *Currency `json:"base_currency,omitempty"`
	TargetCurrency   *Currency `json:"target_currency,omitempty"`
}

type Category struct {
	ID          int       `json:"id"`
	UserID      *int      `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "income" или "expense"
	CreatedAt   time.Time `json:"created_at"`
}

type Transaction struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	CategoryID  int       `json:"category_id"`
	AccountID   *int      `json:"account_id,omitempty"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // "income" или "expense"
	CreatedAt   time.Time `json:"created_at"`
	Account     *Account  `json:"account,omitempty"`
	Category    *Category `json:"category,omitempty"`
}

type Session struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// DTO (Data Transfer Objects)
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type TransactionRequest struct {
	CategoryID  int     `json:"category_id" binding:"required"`
	AccountID   *int    `json:"account_id,omitempty"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
}

// Модель бюджета
type Budget struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	CategoryID int       `json:"category_id"`
	Amount     float64   `json:"amount"`
	Month      string    `json:"month"` // "2025-10"
	CreatedAt  time.Time `json:"created_at"`
}

// МОДЕЛИ ДЛЯ СТАТИСТИКИ
type TransactionSummary struct {
	TotalIncome      float64 `json:"total_income"`
	TotalExpense     float64 `json:"total_expense"`
	NetAmount        float64 `json:"net_amount"`
	TransactionCount int     `json:"transaction_count"`
	PeriodStart      string  `json:"period_start"`
	PeriodEnd        string  `json:"period_end"`
}

type CategorySummary struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Type         string  `json:"type"`
	TotalAmount  float64 `json:"total_amount"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
}

type MonthlySummary struct {
	Month        string  `json:"month"` // "2025-01"
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	NetAmount    float64 `json:"net_amount"`
}

// DTO для дашборда
type DashboardResponse struct {
	Summary            *TransactionSummary `json:"summary"`
	ByCategory         []CategorySummary   `json:"by_category"`
	RecentTransactions []Transaction       `json:"recent_transactions"`
}

// DTO для графиков
type ChartData struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor []string  `json:"backgroundColor"`
}

// НОВЫЕ DTO ДЛЯ ВАЛЮТ И СЧЕТОВ
type CurrencyRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Symbol string `json:"symbol" binding:"required"`
}

type AccountRequest struct {
	Name           string  `json:"name" binding:"required"`
	CurrencyID     int     `json:"currency_id" binding:"required"`
	InitialBalance float64 `json:"initial_balance"`
	IsDefault      *bool   `json:"is_default,omitempty"`
}

type SetDefaultCurrencyRequest struct {
	CurrencyID int `json:"currency_id" binding:"required"`
}

type ExchangeRateRequest struct {
	BaseCurrencyID   int     `json:"base_currency_id" binding:"required"`
	TargetCurrencyID int     `json:"target_currency_id" binding:"required"`
	Rate             float64 `json:"rate" binding:"required,gt=0"`
}

type ConvertCurrencyRequest struct {
	FromAccountID int     `json:"from_account_id" binding:"required"`
	ToAccountID   int     `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}

type ConvertSimpleRequest struct {
	FromCurrencyID int     `json:"from_currency_id" binding:"required"`
	ToCurrencyID   int     `json:"to_currency_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
}

type AccountBalance struct {
	AccountID    int     `json:"account_id"`
	CurrencyCode string  `json:"currency_code"`
	Balance      float64 `json:"balance"`
	BalanceInUSD float64 `json:"balance_in_usd,omitempty"`
}

type CategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required,oneof=income expense"`
}
