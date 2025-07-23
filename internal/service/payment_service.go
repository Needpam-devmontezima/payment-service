package service

import (
	"fmt"
	"log"
	"time"
	"context"

	"github.com/danielmoisemontezima/zw-payment-service/internal/model"
	"github.com/danielmoisemontezima/zw-payment-service/internal/core"
	"github.com/danielmoisemontezima/zw-payment-service/internal/ports"
)

type PaymentService struct {
	providerRegistry *core.ProviderRegistry
	repositoryRegistry *core.RepositoryRegistry
}

func NewPaymentService(providerRegistry *core.ProviderRegistry, repositoryRegistry *core.RepositoryRegistry) *PaymentService {
	return &PaymentService{providerRegistry: providerRegistry, repositoryRegistry: repositoryRegistry}
}

func (s *PaymentService) CreatePaymentIntent(ctx context.Context, provider model.PaymentProvider, req model.PaymentIntentRequest) (*model.PaymentIntentResponse, error) {
	processor, err := s.providerRegistry.Get(provider)
	if err != nil {
		return nil, err
	}

	// Select the repo to interact with table Transactions in DB
	transaction, err := s.repositoryRegistry.Get(core.TransactionRepo)
	if err != nil {
		return nil, err
	}

	// Add payment-processor-specific timeout (2 seconds)
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

	pi_response, err := processor.CreatePaymentIntent(ctx, req)
	if err != nil {
		return nil, err
	}
    
	//Save transaction to repo - table transactions
	newPi := model.Transaction{Amount: pi_response.Amount,
		Currency: pi_response.Currency,
		PaymentIntentID: pi_response.ID,
		TxStatus: ports.Pending,
		CustomerID: req.CustomerID,
		SavePaymentMethod: req.RememberMe,
	}

	err = transaction.Create(ctx, &newPi)
	if err != nil {
		return nil, err
	}
	
	return &model.PaymentIntentResponse{
		ID: pi_response.ID,
		Amount: pi_response.Amount,
		Currency: pi_response.Currency,
		Status: pi_response.Status,
		ClientSecret: pi_response.ClientSecret,
	}, nil
}

func (s *PaymentService) ChargeClient(ctx context.Context, provider model.PaymentProvider, req model.PaymentIntentRequest) (*model.PaymentIntentResponse, error) {
	processor, err := s.providerRegistry.Get(provider)
	if err != nil {
		return nil, err
	}

	// Select the repo to interact with table Transactions in DB
	transaction, err := s.repositoryRegistry.Get(core.TransactionRepo)
	if err != nil {
		return nil, err
	}

	// Select the repo to interact with table payment_method
	payment_method, err := s.repositoryRegistry.Get(core.PaymentRepo)
	if err != nil {
		return nil, err
	}

	// Add payment-processor-specific timeout (5 seconds)
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

	res, err := processor.ChargeClient(ctx, req)
	if err != nil {
		return nil, err
	}

	//Save transaction to repo - table transactions
	newPi := model.Transaction{Amount: res.Amount,
		Currency: res.Currency,
		PaymentIntentID: res.ID,
		TxStatus: ports.PaymentSucceeded,
		CustomerID: req.CustomerID,
		SavePaymentMethod: req.RememberMe,
	}

	err = transaction.Create(ctx, &newPi)
	if err != nil {
		return nil, err
	}

	// Save payment method to repo - table payment_methods
	// Save payment method for this customer if payment is new
	pm_trueVal := true
	if req.RememberMe != nil && *req.RememberMe == pm_trueVal {
		pm, err := payment_method.FindByID(ctx, string(res.PaymentMethodID))
		if err != nil {
			log.Printf("error: %+v", err)
		}

		if pm == nil {
			newPm := model.Transaction{CustomerID: req.CustomerID,
				PaymentMethodID: res.PaymentMethodID,
				PaymentMethodStatus: ports.PaymentMethodStatus,
				PaymentProvider: string(provider),
			}
			err := payment_method.Create(ctx, &newPm)
			if err != nil {
				log.Printf("Saving payment method--Event: %+v", err)
			}
		}
	}

	return &model.PaymentIntentResponse{
		ID: res.ID,
		Amount: res.Amount,
		Currency: res.Currency,
		Status: res.Status,
		ClientSecret: res.ClientSecret,
	}, nil
}

func (s *PaymentService) GetPaymentIntent(ctx context.Context, provider model.PaymentProvider, id string) (*model.PaymentIntentResponse, error) {
	processor, err := s.providerRegistry.Get(provider)
	if err != nil {
		return nil, err
	}

	// Add payment-processor-specific timeout (2 seconds)
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

	pi_response, err := processor.GetPaymentIntent(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.PaymentIntentResponse{
		ID: pi_response.ID,
		Amount: pi_response.Amount,
		Currency: pi_response.Currency,
		Status: pi_response.Status,
		ClientSecret: pi_response.ClientSecret,
	}, nil
}

func (s *PaymentService) ParseWebhook(ctx context.Context, provider model.PaymentProvider, raw []byte, headers map[string][]string) (*model.PaymentEvent, error) {
	processor, err := s.providerRegistry.Get(provider)
	if err != nil {
		return nil, err
	}

	//Select the repo to interact with table Transactions in DB
	transaction, err := s.repositoryRegistry.Get(core.TransactionRepo)
	if err != nil {
		return nil, err
	}

	payment_method, err := s.repositoryRegistry.Get(core.PaymentRepo)
	if err != nil {
		return nil, err
	}

	// Add payment-processor-specific timeout (5 seconds)
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

	event, err := processor.ParseWebhook(ctx, raw, headers)
	if err != nil {
		return nil, err
	}

	switch event.Type {
	case ports.PaymentSucceeded:
		// Handle successful payment
		err := transaction.UpdateStatus(ctx, ports.PaymentSucceeded, map[string]interface{}{
    		"payment_intent_id": event.PaymentIntent,
		})

		if err != nil {
			return nil, err
		}

		// Save payment method if option true
		pm_saveTrue := true
		txdata, err := transaction.FindByPaymentIntent(ctx, string(event.PaymentIntent))

		if err != nil {
			return nil, fmt.Errorf("Error while finding payment intent: %w", err)
		}

		if txdata != nil {
			if txdata.SavePaymentMethod !=nil && *txdata.SavePaymentMethod == pm_saveTrue {
				// Payment method is requested to be saved
				pm, _ := payment_method.FindByID(ctx, string(event.PaymentMethod))
				if pm == nil {
					// Add new payment method to DB
					newPm := model.Transaction{CustomerID: txdata.CustomerID,
						PaymentMethodID: event.PaymentMethod,
						PaymentMethodStatus: ports.PaymentMethodStatus,
						PaymentProvider: string(provider),
					}
					err := payment_method.Create(ctx, &newPm)
					if err != nil {
						log.Printf("Saving payment method--Event: %+v", err)
					}
				}
			}
		}
	case ports.PaymentFailed:
		// Handle failed payment
			err := transaction.UpdateStatus(ctx, ports.PaymentFailed, map[string]interface{}{
    		"payment_intent_id": event.PaymentIntent,
		})

		if err != nil {
			return nil, err
		}
	case ports.PaymentCancelled:
		// Handle cancelled payment
			err := transaction.UpdateStatus(ctx, ports.PaymentCancelled, map[string]interface{}{
    		"payment_intent_id": event.PaymentIntent,
		})

		if err != nil {
			return nil, err
		}
	}

	return event, nil
}

func (s *PaymentService) GetUserPMethods(ctx context.Context, userId string) ([]model.PaymentMethodResponse, error) {
    // Select the repo to interact with table payment_method
    payment_method, err := s.repositoryRegistry.Get(core.PaymentRepo)
    if err != nil {
        return nil, err
    }

    // Get payment methods for the user
    paymentMethods, err := payment_method.FindByColumn(ctx, "customer_id", userId)
    if err != nil {
        log.Printf("Error fetching payment methods: %v", err)
        return nil, fmt.Errorf("failed to get payment methods: %w", err)
    }

    // Transform each payment method to PaymentIntentResponse
    var responses []model.PaymentMethodResponse
    for _, pm := range paymentMethods {
        response := model.PaymentMethodResponse{
            ClientID:				pm.CustomerID,               // Assuming pm has ID field
            PaymentMethodID:  		pm.PaymentMethodID,           // Assuming pm has Amount field
            PaymentProvider:  		pm.PaymentProvider,         // Assuming pm has Currency field
            PaymentMethodType:  	pm.PaymentMethodType,         // Assuming status field is named TxStatus
            PaymentMethodStatus:	pm.PaymentMethodStatus,             // Using userId as ClientSecret as in your example
        }
        responses = append(responses, response)
    }
    return responses, nil
}