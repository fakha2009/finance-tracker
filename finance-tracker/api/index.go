package handler

import (
	"log"
	"net/http"

	"personal-finance-tracker/internal/config"
	"personal-finance-tracker/internal/handler"
	"personal-finance-tracker/internal/repository"
	"personal-finance-tracker/internal/service"
	"personal-finance-tracker/internal/utils"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func init() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.Env != "" {
		gin.SetMode(cfg.Env)
	}

	utils.SetJWTSecret(cfg.JWTSecret)

	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	userService := service.NewUserService(repo)
	transactionService := service.NewTransactionService(repo)
	categoryService := service.NewCategoryService(repo)
	currencyService := service.NewCurrencyService(repo)
	accountService := service.NewAccountService(repo)
	exchangeService := service.NewExchangeService(repo, cfg.ExchangeAPIEndpoint)

	handlers := handler.NewHandler(
		userService,
		transactionService,
		categoryService,
		currencyService,
		accountService,
		exchangeService,
	)

	router = gin.New()
	router.Use(gin.Logger())
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
	}))

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("failed to set trusted proxies: %v", err)
	}

	router.Use(handler.CORSMiddleware(cfg.AllowedOrigin))
	handlers.InitRoutes(router, cfg.CronSecret)

	// Match the standalone server behavior so exchange-dependent screens work
	// on first deploy without waiting for a manual warm-up.
	if err := exchangeService.UpdateExchangeRates(); err != nil {
		log.Printf("initial exchange rates update failed: %v", err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
