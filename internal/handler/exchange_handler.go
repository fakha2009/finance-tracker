package handler

import (
	"net/http"
	"personal-finance-tracker/internal/middleware"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExchangeHandler struct {
	exchangeService service.ExchangeService
	accountService  service.AccountService
}

// ConvertSimpleCurrency конвертирует сумму между валютами по их ID (без привязки к счетам)
func (h *ExchangeHandler) ConvertSimpleCurrency(c *gin.Context) {
	var req models.ConvertSimpleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем курс обмена
	rate, err := h.exchangeService.GetExchangeRate(req.FromCurrencyID, req.ToCurrencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if rate == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "exchange rate not available"})
		return
	}

	converted := req.Amount * rate.Rate
	c.JSON(http.StatusOK, gin.H{
		"from_currency_id": req.FromCurrencyID,
		"to_currency_id":   req.ToCurrencyID,
		"amount":           req.Amount,
		"exchange_rate":    rate.Rate,
		"converted_amount": converted,
	})
}

func NewExchangeHandler(exchangeService service.ExchangeService, accountService service.AccountService) *ExchangeHandler {
	return &ExchangeHandler{
		exchangeService: exchangeService,
		accountService:  accountService,
	}
}

// GetExchangeRates возвращает все курсы валют
func (h *ExchangeHandler) GetExchangeRates(c *gin.Context) {
	rates, err := h.exchangeService.GetAllExchangeRates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rates)
}

// UpdateExchangeRates обновляет курсы валют
func (h *ExchangeHandler) UpdateExchangeRates(c *gin.Context) {
	if err := h.exchangeService.UpdateExchangeRates(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange rates updated successfully"})
}

// ConvertCurrency конвертирует сумму между счетами
func (h *ExchangeHandler) ConvertCurrency(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req models.ConvertCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что счета принадлежат пользователю
	fromAccount, err := h.accountService.GetAccountByID(c.Request.Context(), req.FromAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if fromAccount == nil || fromAccount.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Source account not found or access denied"})
		return
	}

	toAccount, err := h.accountService.GetAccountByID(c.Request.Context(), req.ToAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if toAccount == nil || toAccount.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Target account not found or access denied"})
		return
	}

	// Конвертируем сумму
	convertedAmount, err := h.exchangeService.ConvertAmount(req.Amount, req.FromAccountID, req.ToAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from_account_id":  req.FromAccountID,
		"to_account_id":    req.ToAccountID,
		"original_amount":  req.Amount,
		"converted_amount": convertedAmount,
		"from_currency":    fromAccount.Currency.Code,
		"to_currency":      toAccount.Currency.Code,
	})
}

// GetUserBalances возвращает балансы пользователя в USD
func (h *ExchangeHandler) GetUserBalances(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	balances, err := h.exchangeService.GetUserBalancesInUSD(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balances)
}

// GetExchangeRate возвращает курс между двумя валютами
func (h *ExchangeHandler) GetExchangeRate(c *gin.Context) {
	baseStr := c.Param("base")
	targetStr := c.Param("target")

	baseID, err := strconv.Atoi(baseStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base currency ID"})
		return
	}

	targetID, err := strconv.Atoi(targetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target currency ID"})
		return
	}

	rate, err := h.exchangeService.GetExchangeRate(baseID, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rate == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exchange rate not found"})
		return
	}

	c.JSON(http.StatusOK, rate)
}
