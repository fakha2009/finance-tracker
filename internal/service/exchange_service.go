package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/repository"
	"time"
)

type ExchangeService interface {
	GetExchangeRate(baseCurrencyID, targetCurrencyID int) (*models.ExchangeRate, error)
	GetAllExchangeRates() ([]models.ExchangeRate, error)
	UpdateExchangeRates() error
	ConvertAmount(amount float64, fromAccountID, toAccountID int) (float64, error)
	GetUserBalancesInUSD(userID int) ([]models.AccountBalance, error)
}

type exchangeService struct {
	repo        repository.Repository
	apiEndpoint string
}

// ExchangeRateAPIResponse структура для ответа от ExchangeRate-API
type ExchangeRateAPIResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func NewExchangeService(repo repository.Repository, apiEndpoint string) ExchangeService {
	// fallback на дефолт, если в конфиге пусто
	if apiEndpoint == "" {
		apiEndpoint = "https://api.exchangerate-api.com/v4/latest/USD"
	}
	return &exchangeService{
		repo:        repo,
		apiEndpoint: apiEndpoint,
	}
}

func (s *exchangeService) GetExchangeRate(baseCurrencyID, targetCurrencyID int) (*models.ExchangeRate, error) {
	if baseCurrencyID == targetCurrencyID {
		return &models.ExchangeRate{
			BaseCurrencyID:   baseCurrencyID,
			TargetCurrencyID: targetCurrencyID,
			Rate:             1.0,
			LastUpdated:      time.Now(),
		}, nil
	}

	rate, err := s.repo.GetExchangeRate(baseCurrencyID, targetCurrencyID)
	if err != nil {
		return nil, err
	}

	// Если курс устарел (старше 24 часов), обновляем его
	if rate == nil || time.Since(rate.LastUpdated) > 24*time.Hour {
		err = s.UpdateExchangeRates()
		if err != nil {
			log.Printf("Failed to update exchange rates: %v", err)
			// Возвращаем старый курс, если обновление не удалось
			if rate != nil {
				return rate, nil
			}
		} else {
			// Пытаемся получить обновленный прямой курс
			if r2, e2 := s.repo.GetExchangeRate(baseCurrencyID, targetCurrencyID); e2 == nil && r2 != nil {
				rate = r2
			}
		}
	}

	if rate != nil {
		return rate, nil
	}

	// 1) Пытаемся использовать обратный курс (target->base) как 1/reverse
	if reverse, e := s.repo.GetExchangeRate(targetCurrencyID, baseCurrencyID); e == nil && reverse != nil {
		return &models.ExchangeRate{
			BaseCurrencyID:   baseCurrencyID,
			TargetCurrencyID: targetCurrencyID,
			Rate:             1.0 / reverse.Rate,
			LastUpdated:      reverse.LastUpdated,
			BaseCurrency:     reverse.TargetCurrency,
			TargetCurrency:   reverse.BaseCurrency,
		}, nil
	}

	// 2) Пытаемся вычислить через USD-пивот: base->USD и USD->target
	usd, e := s.repo.GetCurrencyByCode("USD")
	if e != nil {
		return nil, e
	}
	if usd != nil {
		// base -> USD (прямой или обратный)
		var baseToUSD *models.ExchangeRate
		if r, e := s.repo.GetExchangeRate(baseCurrencyID, usd.ID); e == nil {
			baseToUSD = r
		}
		if baseToUSD == nil {
			if r, e := s.repo.GetExchangeRate(usd.ID, baseCurrencyID); e == nil && r != nil {
				baseToUSD = &models.ExchangeRate{Rate: 1.0 / r.Rate, LastUpdated: r.LastUpdated}
			}
		}

		// USD -> target (прямой или обратный)
		var usdToTarget *models.ExchangeRate
		if r, e := s.repo.GetExchangeRate(usd.ID, targetCurrencyID); e == nil {
			usdToTarget = r
		}
		if usdToTarget == nil {
			if r, e := s.repo.GetExchangeRate(targetCurrencyID, usd.ID); e == nil && r != nil {
				usdToTarget = &models.ExchangeRate{Rate: 1.0 / r.Rate, LastUpdated: r.LastUpdated}
			}
		}

		if baseToUSD != nil && usdToTarget != nil {
			// Берем минимально свежее время как lastUpdated
			last := baseToUSD.LastUpdated
			if usdToTarget.LastUpdated.Before(last) {
				last = usdToTarget.LastUpdated
			}
			return &models.ExchangeRate{
				BaseCurrencyID:   baseCurrencyID,
				TargetCurrencyID: targetCurrencyID,
				Rate:             baseToUSD.Rate * usdToTarget.Rate,
				LastUpdated:      last,
			}, nil
		}
	}

	return nil, errors.New("exchange rate not available")
}

func (s *exchangeService) GetAllExchangeRates() ([]models.ExchangeRate, error) {
	rates, err := s.repo.GetAllExchangeRates()
	if err != nil {
		return nil, err
	}

	// Проверяем, есть ли актуальные курсы
	needsUpdate := len(rates) == 0
	if !needsUpdate {
		for _, rate := range rates {
			if time.Since(rate.LastUpdated) > 24*time.Hour {
				needsUpdate = true
				break
			}
		}
	}

	if needsUpdate {
		err = s.UpdateExchangeRates()
		if err != nil {
			log.Printf("Failed to update exchange rates: %v", err)
			// Возвращаем старые курсы, если обновление не удалось
			if len(rates) > 0 {
				return rates, nil
			}
			return nil, errors.New("no exchange rates available")
		}
		// Получаем обновленные курсы
		return s.repo.GetAllExchangeRates()
	}

	return rates, nil
}

func (s *exchangeService) UpdateExchangeRates() error {
	// Получаем все валюты
	currencies, err := s.repo.GetAllCurrencies()
	if err != nil {
		return err
	}

	if len(currencies) == 0 {
		return errors.New("no currencies found")
	}

	// Находим USD как базовую валюту
	var usdCurrency *models.Currency
	for _, currency := range currencies {
		if currency.Code == "USD" {
			usdCurrency = &currency
			break
		}
	}

	if usdCurrency == nil {
		return errors.New("USD currency not found")
	}

	// Получаем курсы из API
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(s.apiEndpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch exchange rates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var apiResponse ExchangeRateAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Сохраняем курсы в базу
	for _, targetCurrency := range currencies {
		if targetCurrency.Code == "USD" {
			continue // Пропускаем базовую валюту
		}

		rate, exists := apiResponse.Rates[targetCurrency.Code]
		if !exists {
			log.Printf("Exchange rate for %s not found in API response", targetCurrency.Code)
			continue
		}

		exchangeRate := &models.ExchangeRate{
			BaseCurrencyID:   usdCurrency.ID,
			TargetCurrencyID: targetCurrency.ID,
			Rate:             rate,
		}

		err = s.repo.CreateOrUpdateExchangeRate(exchangeRate)
		if err != nil {
			log.Printf("Failed to save exchange rate for %s: %v", targetCurrency.Code, err)
		}
	}

	return nil
}

func (s *exchangeService) ConvertAmount(amount float64, fromAccountID, toAccountID int) (float64, error) {
	fromAccount, err := s.repo.GetAccountByID(context.Background(), fromAccountID)
	if err != nil {
		return 0, err
	}
	if fromAccount == nil {
		return 0, errors.New("source account not found")
	}

	toAccount, err := s.repo.GetAccountByID(context.Background(), toAccountID)
	if err != nil {
		return 0, err
	}
	if toAccount == nil {
		return 0, errors.New("target account not found")
	}

	// Если счета в одной валюте, конвертация не нужна
	if fromAccount.CurrencyID == toAccount.CurrencyID {
		return amount, nil
	}

	// Получаем курс обмена
	exchangeRate, err := s.GetExchangeRate(fromAccount.CurrencyID, toAccount.CurrencyID)
	if err != nil {
		return 0, err
	}
	if exchangeRate == nil {
		return 0, errors.New("exchange rate not available")
	}

	convertedAmount := amount * exchangeRate.Rate
	return convertedAmount, nil
}

func (s *exchangeService) GetUserBalancesInUSD(userID int) ([]models.AccountBalance, error) {
	accounts, err := s.repo.GetAccountsByUserID(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	var balances []models.AccountBalance
	usdCurrency, err := s.repo.GetCurrencyByCode("USD")
	if err != nil {
		return nil, err
	}
	if usdCurrency == nil {
		return nil, errors.New("USD currency not found")
	}

	for _, account := range accounts {
		balance := models.AccountBalance{
			AccountID:    account.ID,
			CurrencyCode: account.Currency.Code,
			Balance:      account.Balance,
		}

		// Конвертируем баланс в USD
		if account.Currency.Code != "USD" {
			exchangeRate, err := s.GetExchangeRate(account.CurrencyID, usdCurrency.ID)
			if err == nil && exchangeRate != nil {
				balance.BalanceInUSD = account.Balance * exchangeRate.Rate
			}
		} else {
			balance.BalanceInUSD = account.Balance
		}

		balances = append(balances, balance)
	}

	return balances, nil
}
