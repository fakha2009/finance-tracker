package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"personal-finance-tracker/internal/config"
	"personal-finance-tracker/internal/handler"
	"personal-finance-tracker/internal/repository"
	"personal-finance-tracker/internal/service"
	"personal-finance-tracker/internal/utils"
	"syscall"
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
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
	}))

	// Явно запрещаем доверие к любым прокси по умолчанию, чтобы убрать предупреждение Gin
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// Middleware: CORS для доступа фронтенда
	router.Use(handler.CORSMiddleware(cfg.AllowedOrigin))

	// Routes
	handlers.InitRoutes(router, cfg.CronSecret)

	// Perform initial exchange rates update, but DO NOT run a background infinite ticker
	log.Println("Starting initial exchange rates update...")
	if err := exchangeService.UpdateExchangeRates(); err != nil {
		log.Printf("Failed to update exchange rates initially: %v", err)
	} else {
		log.Println("Initial exchange rates updated successfully.")
	}

	// Настройка HTTP сервера с таймаутами для продакшена
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в горутине, чтобы можно было поймать сигнал
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Ожидание сигнала прерывания для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Server exiting")
}
