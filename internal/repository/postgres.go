package repository

import (
	"context"
	"errors"
	"personal-finance-tracker/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Оптимизация настроек connection pool для production
	config.MaxConns = 25                        // Максимальное количество соединений
	config.MinConns = 5                         // Минимальное количество соединений
	config.MaxConnLifetime = time.Hour          // Время жизни соединения
	config.HealthCheckPeriod = 30 * time.Second // Период health check
	config.MaxConnIdleTime = 30 * time.Minute   // Время простоя соединения

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{db: pool}, nil
}

func (r *PostgresRepository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

// User methods
func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return r.db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		string(hashedPassword),
		time.Now(),
		time.Now(),
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Category methods
func (r *PostgresRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (user_id, name, description, type, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		category.UserID,
		category.Name,
		category.Description,
		category.Type,
		time.Now(),
	).Scan(&category.ID, &category.CreatedAt)
}

func (r *PostgresRepository) GetCategoriesByUserID(ctx context.Context, userID int) ([]models.Category, error) {
	query := `
        SELECT id, user_id, name, description, type, created_at
        FROM categories 
        WHERE user_id = $1 OR user_id IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID,
			&category.UserID,
			&category.Name,
			&category.Description,
			&category.Type,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *PostgresRepository) GetCategoryByID(ctx context.Context, id int) (*models.Category, error) {
	query := `
		SELECT id, user_id, name, description, type, created_at
		FROM categories WHERE id = $1
	`

	var category models.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.Description,
		&category.Type,
		&category.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &category, nil
}

// Transaction methods
func (r *PostgresRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (user_id, category_id, account_id, amount, description, date, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		transaction.UserID,
		transaction.CategoryID,
		transaction.AccountID,
		transaction.Amount,
		transaction.Description,
		transaction.Date,
		transaction.Type,
		time.Now(),
	).Scan(&transaction.ID, &transaction.CreatedAt)
}

func (r *PostgresRepository) GetTransactionsByUserID(ctx context.Context, userID int) ([]models.Transaction, error) {
	query := `
        SELECT t.id, t.user_id, t.category_id, t.account_id, t.amount, t.description, t.date, t.type, t.created_at
        FROM transactions t
		WHERE t.user_id = $1
		ORDER BY t.date DESC, t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.CategoryID,
			&transaction.AccountID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Date,
			&transaction.Type,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *PostgresRepository) GetTransactionByID(ctx context.Context, id int) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, account_id, amount, description, date, type, created_at
		FROM transactions WHERE id = $1
	`

	var transaction models.Transaction
	err := r.db.QueryRow(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.CategoryID,
		&transaction.AccountID,
		&transaction.Amount,
		&transaction.Description,
		&transaction.Date,
		&transaction.Type,
		&transaction.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *PostgresRepository) GetTransactionsByPeriod(ctx context.Context, userID int, start, end time.Time) ([]models.Transaction, error) {
	query := `
        SELECT t.id, t.user_id, t.category_id, t.account_id, t.amount, t.description, t.date, t.type, t.created_at
        FROM transactions t
		WHERE t.user_id = $1 AND t.date BETWEEN $2 AND $3
		ORDER BY t.date DESC, t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.CategoryID,
			&transaction.AccountID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Date,
			&transaction.Type,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// Session methods
func (r *PostgresRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		session.UserID,
		session.Token,
		session.ExpiresAt,
		time.Now(),
	).Scan(&session.ID, &session.CreatedAt)
}

func (r *PostgresRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions WHERE token = $1 AND expires_at > $2
	`

	var session models.Session
	err := r.db.QueryRow(ctx, query, token, time.Now()).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *PostgresRepository) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.Exec(context.Background(), query, token)
	return err
}

// User methods
func (r *PostgresRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, updated_at = $3, default_currency_id = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		user.Username,
		user.Email,
		time.Now(),
		user.DefaultCurrencyID,
		user.ID,
	)
	return err
}

func (r *PostgresRepository) SetUserDefaultCurrency(userID, currencyID int) error {
	query := `UPDATE users SET default_currency_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(context.Background(), query, currencyID, time.Now(), userID)
	return err
}

// Currency methods
func (r *PostgresRepository) CreateCurrency(currency *models.Currency) error {
	query := `
		INSERT INTO currencies (code, name, symbol, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		context.Background(),
		query,
		currency.Code,
		currency.Name,
		currency.Symbol,
		time.Now(),
	).Scan(&currency.ID, &currency.CreatedAt)
}

func (r *PostgresRepository) GetAllCurrencies() ([]models.Currency, error) {
	query := `SELECT id, code, name, symbol, created_at FROM currencies ORDER BY code`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []models.Currency
	for rows.Next() {
		var currency models.Currency
		err := rows.Scan(
			&currency.ID,
			&currency.Code,
			&currency.Name,
			&currency.Symbol,
			&currency.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

func (r *PostgresRepository) GetCurrencyByID(ctx context.Context, id int) (*models.Currency, error) {
	query := `SELECT id, code, name, symbol, created_at FROM currencies WHERE id = $1`

	var currency models.Currency
	err := r.db.QueryRow(ctx, query, id).Scan(
		&currency.ID,
		&currency.Code,
		&currency.Name,
		&currency.Symbol,
		&currency.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &currency, nil
}

func (r *PostgresRepository) GetCurrencyByCode(code string) (*models.Currency, error) {
	query := `SELECT id, code, name, symbol, created_at FROM currencies WHERE code = $1`

	var currency models.Currency
	err := r.db.QueryRow(context.Background(), query, code).Scan(
		&currency.ID,
		&currency.Code,
		&currency.Name,
		&currency.Symbol,
		&currency.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &currency, nil
}

// Account methods
func (r *PostgresRepository) CreateAccount(account *models.Account) error {
	query := `
		INSERT INTO accounts (user_id, currency_id, balance, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		context.Background(),
		query,
		account.UserID,
		account.CurrencyID,
		account.Balance,
		account.IsDefault,
		time.Now(),
		time.Now(),
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *PostgresRepository) GetAccountsByUserID(ctx context.Context, userID int) ([]models.Account, error) {
	query := `
		SELECT a.id, a.user_id, a.currency_id, a.balance, a.is_default, a.created_at, a.updated_at,
		       c.id, c.code, c.name, c.symbol, c.created_at
		FROM accounts a
		JOIN currencies c ON a.currency_id = c.id
		WHERE a.user_id = $1
		ORDER BY a.is_default DESC, a.created_at
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		var currency models.Currency
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.CurrencyID,
			&account.Balance,
			&account.IsDefault,
			&account.CreatedAt,
			&account.UpdatedAt,
			&currency.ID,
			&currency.Code,
			&currency.Name,
			&currency.Symbol,
			&currency.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		account.Currency = &currency
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *PostgresRepository) GetAccountByID(ctx context.Context, id int) (*models.Account, error) {
	query := `
		SELECT a.id, a.user_id, a.currency_id, a.balance, a.is_default, a.created_at, a.updated_at,
		       c.id, c.code, c.name, c.symbol, c.created_at
		FROM accounts a
		JOIN currencies c ON a.currency_id = c.id
		WHERE a.id = $1
	`

	var account models.Account
	var currency models.Currency
	err := r.db.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.CurrencyID,
		&account.Balance,
		&account.IsDefault,
		&account.CreatedAt,
		&account.UpdatedAt,
		&currency.ID,
		&currency.Code,
		&currency.Name,
		&currency.Symbol,
		&currency.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	account.Currency = &currency
	return &account, nil
}

func (r *PostgresRepository) GetDefaultAccount(ctx context.Context, userID int) (*models.Account, error) {
	query := `
		SELECT a.id, a.user_id, a.currency_id, a.balance, a.is_default, a.created_at, a.updated_at,
		       c.id, c.code, c.name, c.symbol, c.created_at
		FROM accounts a
		JOIN currencies c ON a.currency_id = c.id
		WHERE a.user_id = $1 AND a.is_default = true
	`

	var account models.Account
	var currency models.Currency
	err := r.db.QueryRow(context.Background(), query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.CurrencyID,
		&account.Balance,
		&account.IsDefault,
		&account.CreatedAt,
		&account.UpdatedAt,
		&currency.ID,
		&currency.Code,
		&currency.Name,
		&currency.Symbol,
		&currency.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	account.Currency = &currency
	return &account, nil
}

func (r *PostgresRepository) UpdateAccountBalance(accountID int, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(context.Background(), query, newBalance, time.Now(), accountID)
	return err
}

func (r *PostgresRepository) SetDefaultAccount(userID, accountID int) error {
	// Сначала сбрасываем все счета пользователя как не-дефолтные
	resetQuery := `UPDATE accounts SET is_default = false WHERE user_id = $1`
	_, err := r.db.Exec(context.Background(), resetQuery, userID)
	if err != nil {
		return err
	}

	// Устанавливаем выбранный счет как дефолтный
	setQuery := `UPDATE accounts SET is_default = true, updated_at = $1 WHERE id = $2 AND user_id = $3`
	_, err = r.db.Exec(context.Background(), setQuery, time.Now(), accountID, userID)
	return err
}

// Exchange Rate methods
func (r *PostgresRepository) CreateOrUpdateExchangeRate(rate *models.ExchangeRate) error {
	query := `
		INSERT INTO exchange_rates (base_currency_id, target_currency_id, rate, last_updated)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (base_currency_id, target_currency_id) 
		DO UPDATE SET rate = EXCLUDED.rate, last_updated = EXCLUDED.last_updated
		RETURNING id, last_updated
	`

	return r.db.QueryRow(
		context.Background(),
		query,
		rate.BaseCurrencyID,
		rate.TargetCurrencyID,
		rate.Rate,
		time.Now(),
	).Scan(&rate.ID, &rate.LastUpdated)
}

func (r *PostgresRepository) GetExchangeRate(baseCurrencyID, targetCurrencyID int) (*models.ExchangeRate, error) {
	query := `
		SELECT er.id, er.base_currency_id, er.target_currency_id, er.rate, er.last_updated,
		       bc.id, bc.code, bc.name, bc.symbol, bc.created_at,
		       tc.id, tc.code, tc.name, tc.symbol, tc.created_at
		FROM exchange_rates er
		JOIN currencies bc ON er.base_currency_id = bc.id
		JOIN currencies tc ON er.target_currency_id = tc.id
		WHERE er.base_currency_id = $1 AND er.target_currency_id = $2
	`

	var rate models.ExchangeRate
	var baseCurrency, targetCurrency models.Currency
	err := r.db.QueryRow(context.Background(), query, baseCurrencyID, targetCurrencyID).Scan(
		&rate.ID,
		&rate.BaseCurrencyID,
		&rate.TargetCurrencyID,
		&rate.Rate,
		&rate.LastUpdated,
		&baseCurrency.ID,
		&baseCurrency.Code,
		&baseCurrency.Name,
		&baseCurrency.Symbol,
		&baseCurrency.CreatedAt,
		&targetCurrency.ID,
		&targetCurrency.Code,
		&targetCurrency.Name,
		&targetCurrency.Symbol,
		&targetCurrency.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rate.BaseCurrency = &baseCurrency
	rate.TargetCurrency = &targetCurrency
	return &rate, nil
}

func (r *PostgresRepository) GetAllExchangeRates() ([]models.ExchangeRate, error) {
	query := `
		SELECT er.id, er.base_currency_id, er.target_currency_id, er.rate, er.last_updated,
		       bc.id, bc.code, bc.name, bc.symbol, bc.created_at,
		       tc.id, tc.code, tc.name, tc.symbol, tc.created_at
		FROM exchange_rates er
		JOIN currencies bc ON er.base_currency_id = bc.id
		JOIN currencies tc ON er.target_currency_id = tc.id
		ORDER BY bc.code, tc.code
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.ExchangeRate
	for rows.Next() {
		var rate models.ExchangeRate
		var baseCurrency, targetCurrency models.Currency
		err := rows.Scan(
			&rate.ID,
			&rate.BaseCurrencyID,
			&rate.TargetCurrencyID,
			&rate.Rate,
			&rate.LastUpdated,
			&baseCurrency.ID,
			&baseCurrency.Code,
			&baseCurrency.Name,
			&baseCurrency.Symbol,
			&baseCurrency.CreatedAt,
			&targetCurrency.ID,
			&targetCurrency.Code,
			&targetCurrency.Name,
			&targetCurrency.Symbol,
			&targetCurrency.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rate.BaseCurrency = &baseCurrency
		rate.TargetCurrency = &targetCurrency
		rates = append(rates, rate)
	}

	return rates, nil
}

func (r *PostgresRepository) GetExchangeRatesByBaseCurrency(baseCurrencyID int) ([]models.ExchangeRate, error) {
	query := `
		SELECT er.id, er.base_currency_id, er.target_currency_id, er.rate, er.last_updated,
		       bc.id, bc.code, bc.name, bc.symbol, bc.created_at,
		       tc.id, tc.code, tc.name, tc.symbol, tc.created_at
		FROM exchange_rates er
		JOIN currencies bc ON er.base_currency_id = bc.id
		JOIN currencies tc ON er.target_currency_id = tc.id
		WHERE er.base_currency_id = $1
		ORDER BY tc.code
	`

	rows, err := r.db.Query(context.Background(), query, baseCurrencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.ExchangeRate
	for rows.Next() {
		var rate models.ExchangeRate
		var baseCurrency, targetCurrency models.Currency
		err := rows.Scan(
			&rate.ID,
			&rate.BaseCurrencyID,
			&rate.TargetCurrencyID,
			&rate.Rate,
			&rate.LastUpdated,
			&baseCurrency.ID,
			&baseCurrency.Code,
			&baseCurrency.Name,
			&baseCurrency.Symbol,
			&baseCurrency.CreatedAt,
			&targetCurrency.ID,
			&targetCurrency.Code,
			&targetCurrency.Name,
			&targetCurrency.Symbol,
			&targetCurrency.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rate.BaseCurrency = &baseCurrency
		rate.TargetCurrency = &targetCurrency
		rates = append(rates, rate)
	}

	return rates, nil
}

// New method to get transactions by account ID
func (r *PostgresRepository) GetTransactionsByAccountID(ctx context.Context, accountID int) ([]models.Transaction, error) {
	query := `
        SELECT t.id, t.user_id, t.category_id, t.account_id, t.amount, t.description, t.date, t.type, t.created_at
        FROM transactions t
		WHERE t.account_id = $1
		ORDER BY t.date DESC, t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.CategoryID,
			&transaction.AccountID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Date,
			&transaction.Type,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *PostgresRepository) GetTransactionsByAccountIDAndPeriod(ctx context.Context, accountID int, start, end time.Time) ([]models.Transaction, error) {
	query := `
        SELECT t.id, t.user_id, t.category_id, t.account_id, t.amount, t.description, t.date, t.type, t.created_at
        FROM transactions t
		WHERE t.account_id = $1 AND t.date >= $2 AND t.date <= $3
		ORDER BY t.date DESC, t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, accountID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.CategoryID,
			&transaction.AccountID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Date,
			&transaction.Type,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *PostgresRepository) GetTransactionSummaryByAccountID(ctx context.Context, accountID int, start, end time.Time) (*models.TransactionSummary, error) {
	query := `
        SELECT 
            COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) as total_income,
            COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) as total_expense,
            COUNT(t.id) as transaction_count
        FROM transactions t
        WHERE t.account_id = $1 AND t.date >= $2 AND t.date <= $3
	`

	var totalIncome, totalExpense float64
	var transactionCount int
	err := r.db.QueryRow(ctx, query, accountID, start, end).Scan(&totalIncome, &totalExpense, &transactionCount)
	if err != nil {
		return nil, err
	}

	return &models.TransactionSummary{
		TotalIncome:      totalIncome,
		TotalExpense:     totalExpense,
		NetAmount:        totalIncome - totalExpense,
		TransactionCount: transactionCount,
		PeriodStart:      start.Format("2006-01-02"),
		PeriodEnd:        end.Format("2006-01-02"),
	}, nil
}
