package adapters

import (
    "errors"
    "context"
    "encoding/json"
    
    "github.com/stripe/stripe-go/v72"
    "github.com/stripe/stripe-go/v72/webhook"
    "github.com/stripe/stripe-go/v72/customer"
    "github.com/stripe/stripe-go/v72/paymentintent"
    "github.com/stripe/stripe-go/v72/paymentmethod"
    "github.com/danielmoisemontezima/zw-payment-service/pkg/utils"
    "github.com/danielmoisemontezima/zw-payment-service/internal/ports"
    "github.com/danielmoisemontezima/zw-payment-service/internal/model"
)

type StripeAdapter struct {
    apiKey string
    webhookSecret string
}

func NewStripeAdapter(apiKey string, webhookSecret string) *StripeAdapter {
    stripe.Key = apiKey
    return &StripeAdapter{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
	}
}

func (s *StripeAdapter) Name() model.PaymentProvider {
    var ppStripe model.PaymentProvider = "stripe"
    return ppStripe
}

func (s *StripeAdapter) CreatePaymentIntent(ctx context.Context, req model.PaymentIntentRequest) (*model.PaymentProcessorResponse, error) {
    params := &stripe.PaymentIntentParams{
        Amount:   stripe.Int64(req.Amount),
        Currency: stripe.String(req.Currency),
    }
    
    pi, err := paymentintent.New(params)
    if err != nil {
        return nil, errors.New("Payment creation failed #acpi0")
    }
    
    return &model.PaymentProcessorResponse{
        ID:           pi.ID,
        Amount:       pi.Amount,
        Currency:     string(pi.Currency),
        Status:       string(pi.Status),
        ClientSecret: pi.ClientSecret,
    }, nil
}

func (s *StripeAdapter) GetPaymentIntent(ctx context.Context, id string) (*model.PaymentProcessorResponse, error) {
    pi, err := paymentintent.Get(id, nil)

    if err != nil {
        return nil, errors.New("Collecting payment details failed #agpi0")
    }

    return &model.PaymentProcessorResponse{
        ID:           pi.ID,
        Amount:       pi.Amount,
        Currency:     string(pi.Currency),
        Status:       string(pi.Status),
        ClientSecret: pi.ClientSecret,
    }, nil
}

func (s *StripeAdapter) ChargeClient(ctx context.Context, req model.PaymentIntentRequest) (*model.PaymentProcessorResponse, error) {
    // Check if PaymentMethod is already attached
    pm, err := paymentmethod.Get(string(req.Token), nil)
    if err != nil {
       return nil, errors.New("Direct charge failed #acc0")
    }

    //No customer attached. Attach new customer
    if pm.Customer == nil {
        // Create a Customer
        customerParams := &stripe.CustomerParams{}
        customer, err := customer.New(customerParams)
        if err != nil {
            return nil, errors.New("Direct charge failed #acc1")
        }

        // Attach the PaymentMethod to the Customer
        attachParams := &stripe.PaymentMethodAttachParams{
            Customer: stripe.String(customer.ID),
        }

        _, err = paymentmethod.Attach(string(req.Token), attachParams) // or "pm_123"
        if err != nil {
            return nil, errors.New("Direct charge failed #acc2")
        }

        pm.Customer.ID = customer.ID
    }

    // Use it in a PaymentIntent
    params := &stripe.PaymentIntentParams{
        Amount:   stripe.Int64(req.Amount),
        Currency: stripe.String(req.Currency),
        Customer: stripe.String( pm.Customer.ID),
        PaymentMethod: stripe.String(req.Token),
        Confirm:  stripe.Bool(true),
    }

    pi, err := paymentintent.New(params)
    if err != nil {
        return nil, errors.New("Direct charge failed #acc3")
    }

    return &model.PaymentProcessorResponse{
        ID:           pi.ID,
        Amount:       pi.Amount,
        Currency:     string(pi.Currency),
        Status:       string(pi.Status),
        PaymentMethodID: pi.PaymentMethod.ID,
    }, nil
}

func (s *StripeAdapter) ParseWebhook(ctx context.Context, raw []byte, headers map[string][]string) (*model.PaymentEvent, error) {
	// Extract Stripe-Signature header
	sigHeader := utils.GetHeader(headers, "Stripe-Signature")
	if sigHeader == "" {
		return nil, errors.New("missing stripe-signature header")
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(raw, sigHeader, s.webhookSecret)
	if err != nil {
		return nil, errors.New("invalid webhook signature")
	}

	// Parse event based on type
	switch event.Type {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			return nil, errors.New("failed to parse event data")
		}

		// Get PaymentMethod ID (handles both attached & detached methods)
		paymentMethodID := ""
		if paymentIntent.PaymentMethod != nil {
			paymentMethodID = paymentIntent.PaymentMethod.ID
		} else if paymentIntent.LastPaymentError != nil && paymentIntent.LastPaymentError.PaymentMethod != nil {
			paymentMethodID = paymentIntent.LastPaymentError.PaymentMethod.ID
		}

		// Determine event type
		eventType := ports.PaymentSucceeded
		if event.Type == "payment_intent.payment_failed" {
			eventType = ports.PaymentFailed
		} else if event.Type == "payment_intent.canceled" {
			eventType = ports.PaymentCancelled
		}

		return &model.PaymentEvent{
			Type:          eventType,
			PaymentIntent: paymentIntent.ID,
			PaymentMethod: paymentMethodID, // Include Payment Method ID
			Payload: model.PaymentIntentResponse{
				ID:           paymentIntent.ID,
				Amount:       paymentIntent.Amount,
				Currency:     string(paymentIntent.Currency),
				Status:       string(paymentIntent.Status),
			},
		}, nil

	default:
		return &model.PaymentEvent{
			Type:    event.Type,
			Payload: "Omitted",
		}, nil
	}
}