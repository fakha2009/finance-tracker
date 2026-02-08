package handler

import (
	"net/http"
	"personal-finance-tracker/internal/middleware"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	accountService service.AccountService
}

func NewAccountHandler(accountService service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// CreateAccount создает новый счет
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req models.AccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account := &models.Account{
		UserID:     user.ID,
		CurrencyID: req.CurrencyID,
		Balance:    req.InitialBalance,
		IsDefault:  false,
	}

	if req.IsDefault != nil {
		account.IsDefault = *req.IsDefault
	}

	if err := h.accountService.CreateAccount(c.Request.Context(), account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

// GetUserAccounts возвращает все счета пользователя
func (h *AccountHandler) GetUserAccounts(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	accounts, err := h.accountService.GetUserAccounts(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetAccountByID возвращает счет по ID
func (h *AccountHandler) GetAccountByID(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	account, err := h.accountService.GetAccountByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if account == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Проверяем, что счет принадлежит пользователю
	if account.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// SetDefaultAccount устанавливает счет по умолчанию
func (h *AccountHandler) SetDefaultAccount(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	if err := h.accountService.SetDefaultAccount(c.Request.Context(), user.ID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default account updated successfully"})
}
