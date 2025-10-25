package handler

import (
	"net/http"
	"personal-finance-tracker/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CurrencyHandler struct {
	currencyService service.CurrencyService
}

func NewCurrencyHandler(currencyService service.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		currencyService: currencyService,
	}
}

// GetAllCurrencies возвращает все доступные валюты
func (h *CurrencyHandler) GetAllCurrencies(c *gin.Context) {
	currencies, err := h.currencyService.GetAllCurrencies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, currencies)
}

// GetCurrencyByID возвращает валюту по ID
func (h *CurrencyHandler) GetCurrencyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency ID"})
		return
	}

	currency, err := h.currencyService.GetCurrencyByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if currency == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Currency not found"})
		return
	}

	c.JSON(http.StatusOK, currency)
}