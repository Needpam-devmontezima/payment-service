package ports

import (
	"context"

	"github.com/danielmoisemontezima/zw-payment-service/internal/model"
)

const (
	PaymentMethodStatus = "active"
	PaymentSucceeded = "payment_succeeded"
	PaymentFailed    = "payment_failed"
	PaymentCancelled = "payment_intent.canceled"
	Pending			 = "pending"
)

type IPaymentProcessor interface {
	Name() model.PaymentProvider
	CreatePaymentIntent(ctx context.Context, req model.PaymentIntentRequest) (*model.PaymentProcessorResponse, error)
	GetPaymentIntent(ctx context.Context, id string) (*model.PaymentProcessorResponse, error)
	ChargeClient(ctx context.Context, req model.PaymentIntentRequest) (*model.PaymentProcessorResponse, error)
	ParseWebhook(ctx context.Context, raw []byte, headers map[string][]string) (*model.PaymentEvent, error)
}