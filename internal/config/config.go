package config

import (
	"log"
	"os"
	"strconv"
	
	"github.com/joho/godotenv"
)

type Config struct {
	StripeSecretKey      string
	StripeWebhookSecret  string
	PayPalClientID       string
	PayPalSecret         string
	DbUser				 string
	DbPassword			 string
	DbHost				 string
	DbName				 string
	SSLMode				 string
	DbPort				 string
	Port                int
}

func Load() (*Config, error) {
	// Load .env file (only in development)
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}

	return &Config{
		StripeSecretKey:      os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
		PayPalClientID:       os.Getenv("PAYPAL_CLIENT_ID"),
		PayPalSecret:         os.Getenv("PAYPAL_SECRET"),
		DbUser:				  os.Getenv("DB_USER"),
		DbPassword:			  os.Getenv("DB_PASSWORD"),
		DbHost:				  os.Getenv("DB_HOST"),
		DbName:				  os.Getenv("DB_NAME"),
		DbPort:				  os.Getenv("DB_PORT"),
		SSLMode:			  os.Getenv("SSL_MODE"),
		Port:                 port,
	}, nil
}