package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"appa_subscriptions/internal/config"
	"appa_subscriptions/internal/handlers"
	"appa_subscriptions/internal/jobs"
	"appa_subscriptions/internal/routers"
	"appa_subscriptions/internal/services"
	"appa_subscriptions/pkg/db"
	PaymentInstallment "appa_subscriptions/pkg/db/repositories"
	"appa_subscriptions/pkg/logs"
	"appa_subscriptions/pkg/mailgun"
	"appa_subscriptions/pkg/shopify"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.Debug == "" {
		cfg.Debug = "0"
	}

	logger := logs.NewZapLogger()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("error syncing logger: %v\n", err)
		}
	}()

	sslmode := cfg.SSLMode
	fmt.Printf("sslmode -> %s\n", sslmode)
	if len(sslmode) > 0 {
		sslmode = "sslmode=" + sslmode
	}

	//connect the database
	connStr := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s %s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, sslmode)
	// gorm connect
	gormDB, err := db.NewDBSQLHandler(connStr)
	if err != nil {
		logger.Error(err.Error(), zap.Any("host", cfg.DBHost), zap.Any("port", cfg.DBPort), zap.Any("user", cfg.DBUser), zap.Any("dbname", cfg.DBName))
	}

	db, err := gormDB.DB()
	if err != nil {
		logger.Error(err.Error(), zap.Any("host", cfg.DBHost), zap.Any("port", cfg.DBPort), zap.Any("user", cfg.DBUser), zap.Any("dbname", cfg.DBName))
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("error db body: %v\n", err)
		}
	}()

	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		logger.Fatal("could not load Venezuela time zone", zap.Error(err))
	}

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	// Initialize API clients / repositories
	shopifyCliente := shopify.NewRepository(
		cfg.ShopifyStoreName, cfg.ShopifyAPIVersion, cfg.ShopifyAdminToken, logger,
	)
	paymentInstallmentRepo := PaymentInstallment.NewPaymentInstallmentRepository(loc, logger)

	muClient := mailgun.NewClient(cfg.MailgunAPIKey)

	// Initialize resources
	muRepository := mailgun.NewRepository(muClient, cfg.MailgunDomain, cfg.MailgunSender, logger)

	// Initialize services
	webhookService := services.NewWebhookService(gormDB, loc, shopifyCliente, paymentInstallmentRepo, logger)
	orderService := services.NewOrderService(gormDB, shopifyCliente, paymentInstallmentRepo, muRepository, loc, logger)
	services.NewNotificationService(muRepository, logger)

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(webhookService)

	// Initialize routes
	webhookRouter := routers.NewWebhookRoutes(webhookHandler)

	// Set up routes
	webhookRouter.SetRouter(router, cfg.ShopifyHMACSecret)

	// Jobs
	jobHandler := jobs.NewJobHandler(orderService, logger)

	// init config cron
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(loc),
	)

	// Add TIIE job -> RUN | 08:30am | ALL DAYS |
	_, err = c.AddFunc("0 30 8 * * *", jobHandler.HandleScheduledOrders)
	if err != nil {
		logger.Fatal("error adding job HandleScheduledOrders to cron", zap.Error(err))
	}

	// Add TIIE job -> RUN | 09:30am | ALL DAYS |
	_, err = c.AddFunc("0 30 9 * * *", jobHandler.HandleReminderPendingPolicies)
	if err != nil {
		logger.Fatal("error adding job HandleScheduledOrders to cron", zap.Error(err))
	}

	if cfg.Debug != "1" {
		c.Start()
		defer c.Stop()
	}

	// testing run job
	// jobHandler.HandleReminderPendingPolicies()

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
