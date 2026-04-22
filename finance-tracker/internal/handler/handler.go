package handler

import (
	"handler/internal/middleware"
	"handler/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService        service.UserService
	transactionService service.TransactionService
	categoryService    service.CategoryService
	currencyService    service.CurrencyService
	accountService     service.AccountService
	exchangeService    service.ExchangeService
}

func NewHandler(
	userService service.UserService,
	transactionService service.TransactionService,
	categoryService service.CategoryService,
	currencyService service.CurrencyService,
	accountService service.AccountService,
	exchangeService service.ExchangeService,
) *Handler {
	return &Handler{
		userService:        userService,
		transactionService: transactionService,
		categoryService:    categoryService,
		currencyService:    currencyService,
		accountService:     accountService,
		exchangeService:    exchangeService,
	}
}

func (h *Handler) InitRoutes(router *gin.Engine, cronSecret string) {
	// Инициализация отдельных хендлеров
	currencyHandler := NewCurrencyHandler(h.currencyService)
	accountHandler := NewAccountHandler(h.accountService)
	exchangeHandler := NewExchangeHandler(h.exchangeService, h.accountService)

	// Группа публичных маршрутов (не требует аутентификации)
	public := router.Group("/api/v1")
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
		public.GET("/health", h.HealthCheck)
		public.GET("/currencies", currencyHandler.GetAllCurrencies)
		public.GET("/currencies/:id", currencyHandler.GetCurrencyByID)
		// Обмен валют (read-only): доступны без авторизации
		public.GET("/exchange/rates", exchangeHandler.GetExchangeRates)
		public.GET("/exchange/rate/:base/:target", exchangeHandler.GetExchangeRate)
		public.POST("/exchange/convert-simple", exchangeHandler.ConvertSimpleCurrency)
		
		// Cron endpoint 
		cronGroup := public.Group("/cron")
		cronGroup.Use(CronMiddleware(cronSecret))
		cronGroup.POST("/update-rates", exchangeHandler.UpdateExchangeRates)
	}

	// Группа защищенных маршрутов (требуется JWT + активная сессия)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(h.userService))
	{
		// Пользователь
		protected.GET("/user/profile", h.GetUserProfile)
		protected.POST("/logout", h.Logout)
		protected.PUT("/user/default-currency", h.SetDefaultCurrency)
		protected.GET("/user/profile-with-accounts", h.GetUserProfileWithAccounts)

		// Категории
		categoryHandler := NewCategoryHandler(h.categoryService)
		protected.GET("/categories", categoryHandler.GetCategories)
		protected.POST("/categories", categoryHandler.CreateCategory)
		protected.DELETE("/categories/:id", categoryHandler.DeleteCategory)

		// Транзакции
		protected.GET("/transactions", h.GetTransactions)
		protected.POST("/transactions", h.CreateTransaction)
		protected.GET("/transactions/period", h.GetTransactionsByPeriod)
		protected.GET("/transactions/:id", h.GetTransaction)

		// Счета
		protected.GET("/accounts", accountHandler.GetUserAccounts)
		protected.POST("/accounts", accountHandler.CreateAccount)
		protected.GET("/accounts/:id", accountHandler.GetAccountByID)
		protected.PUT("/accounts/:id/default", accountHandler.SetDefaultAccount)

		// Обмен валют
		protected.POST("/exchange/rates/update", exchangeHandler.UpdateExchangeRates)
		protected.POST("/exchange/convert", exchangeHandler.ConvertCurrency)
		protected.GET("/exchange/balances", exchangeHandler.GetUserBalances)

		// Статистика транзакций
		protected.GET("/transactions/summary", h.GetTransactionsSummary)
		protected.GET("/transactions/by-category", h.GetTransactionsByCategory)
		protected.GET("/transactions/monthly-summary", h.GetMonthlySummary)
	}
}

// Cron Middleware
func CronMiddleware(cronSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-CRON-KEY")
		if key != cronSecret || cronSecret == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized cron request"})
			return
		}
		c.Next()
	}
}

// CORS Middleware
func CORSMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Strict CORS checking
		if allowedOrigin == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Set strictly the allowed origin, regardless of requester
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			// Credentials are only allowed for specific origins
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
