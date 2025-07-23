package main

import (
	"fmt"
	"log"
	"strconv"
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"github.com/danielmoisemontezima/zw-payment-service/internal/config"
	"github.com/danielmoisemontezima/zw-payment-service/internal/adapters"
	"github.com/danielmoisemontezima/zw-payment-service/internal/repository"
	"github.com/danielmoisemontezima/zw-payment-service/internal/controller"
	"github.com/danielmoisemontezima/zw-payment-service/internal/core"
	"github.com/danielmoisemontezima/zw-payment-service/internal/model"
	"github.com/danielmoisemontezima/zw-payment-service/internal/service"
)

func main() {
	// Setup
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//Load configurations
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Initialize database pool
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=180",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbName,
		cfg.SSLMode,
	)

	//Debug purposes
	log.Printf("configurations: %s", connString)

	pool, err := config.InitPostgresPool(connString)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}
	defer pool.Close()

	// Declare providers
	const stripe  model.PaymentProvider = "stripe"
	//const paypal model.PaymentProvider = "paypal"

	// Initialize providers
	providerRegistry := core.NewProviderRegistry()
	providerRegistry.Register(stripe, adapters.NewStripeAdapter(cfg.StripeSecretKey, cfg.StripeWebhookSecret))
	//registry.Register(paypal, adapters.NewPayPalAdapter(cfg.PayPalClientID, cfg.PayPalSecret))

	// Initialize repositories
	repositoryRegistry := core.NewRepositoryRegistry()
	repositoryRegistry.Register(core.TransactionRepo, repository.NewTransactionRepository(pool))
	repositoryRegistry.Register(core.PaymentRepo, repository.NewPaymentMethodRepository(pool))

	// Setup services
	paymentService := service.NewPaymentService(providerRegistry, repositoryRegistry)
	paymentController := controller.NewPaymentController(paymentService)

	// Router
	r := chi.NewRouter()
	r.Post("/payments/{provider}/intent", paymentController.CreatePaymentIntent)
	r.Post("/payments/{provider}/charge", paymentController.ChargeClient)
	r.Post("/webhooks/{provider}", paymentController.ParseWebhook)

	r.Get("/payments/health", paymentController.GetHealthCheck)
	r.Get("/payments/{provider}/intent", paymentController.GetPaymentIntent)
	r.Get("/payments/methods/{id}", paymentController.GetUserPMethods)

	// Start server
	log.Printf("configurations: %+v", cfg)
	log.Printf("Server running on :%d", cfg.Port)
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}