package main

import (
	"log"
	"personal-finance-tracker/internal/config"
	"personal-finance-tracker/internal/handler"
	"personal-finance-tracker/internal/repository"
	"personal-finance-tracker/internal/service"
	"personal-finance-tracker/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Устанавливаем режим Gin из конфигурации (debug/release)
	if cfg.Env != "" {
		gin.SetMode(cfg.Env)
	}

	utils.SetJWTSecret(cfg.JWTSecret)

	// Инициализация репозитория
	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer repo.Close()

	// Инициализация сервисов (бизнес-логика)
	userService := service.NewUserService(repo)
	transactionService := service.NewTransactionService(repo)
	categoryService := service.NewCategoryService(repo)
	currencyService := service.NewCurrencyService(repo)
	accountService := service.NewAccountService(repo)
	exchangeService := service.NewExchangeService(repo, cfg.ExchangeAPIEndpoint)

	// Инициализация обработчиков
	handlers := handler.NewHandler(
		userService,
		transactionService,
		categoryService,
		currencyService,
		accountService,
		exchangeService,
	)

	// Настройка роутера
	router := gin.Default()
	// Явно запрещаем доверие к любым прокси по умолчанию, чтобы убрать предупреждение Gin
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// Middleware: CORS для доступа фронтенда
	router.Use(handler.CORSMiddleware())

	// Routes
	handlers.InitRoutes(router)

	// Запускаем фоновое обновление курсов валют при старте
	go func() {
		log.Println("Starting initial exchange rates update...")
		if err := exchangeService.UpdateExchangeRates(); err != nil {
			log.Printf("Failed to update exchange rates: %v", err)
		} else {
			log.Println("Exchange rates updated successfully")
		}

		// Периодическое обновление каждые 12 часов
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("Scheduled exchange rates update...")
			if err := exchangeService.UpdateExchangeRates(); err != nil {
				log.Printf("Failed to update exchange rates on schedule: %v", err)
			} else {
				log.Println("Scheduled exchange rates updated successfully")
			}
		}
	}()

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
