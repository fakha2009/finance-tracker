package handler

import (
	"net/http"
	"personal-finance-tracker/internal/middleware"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTransactionsSummary возвращает сводку по транзакциям
func (h *Handler) GetTransactionsSummary(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
		return
	}

	summary, err := h.transactionService.GetTransactionSummary(c.Request.Context(), user.ID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetTransactionsByCategory возвращает статистику по категориям
func (h *Handler) GetTransactionsByCategory(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
		return
	}

	categorySummary, err := h.transactionService.GetTransactionsByCategory(c.Request.Context(), user.ID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categorySummary)
}

// GetMonthlySummary возвращает месячную статистику
func (h *Handler) GetMonthlySummary(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		year = time.Now().Year()
	}

	monthlySummary, err := h.transactionService.GetMonthlySummary(c.Request.Context(), user.ID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, monthlySummary)
}
