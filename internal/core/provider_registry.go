package core

import (
    "github.com/danielmoisemontezima/zw-payment-service/internal/ports"
    "github.com/danielmoisemontezima/zw-payment-service/internal/model"
    "fmt"
)

const (
	TransactionRepo = "transactionRepo"
	PaymentRepo    = "paymentRepo"
	RefundRepo    = "refundRepo"
)

type RepositoryRegistry struct {
    repo map[string]ports.ITransactionRepository
}

func NewRepositoryRegistry() *RepositoryRegistry {
    return &RepositoryRegistry{
        repo: make(map[string]ports.ITransactionRepository),
    }
}

func (r *RepositoryRegistry) Register(repoTab string, repo ports.ITransactionRepository) {
    r.repo[repoTab] = repo
}

func (r *RepositoryRegistry) Get(tab string) (ports.ITransactionRepository, error) {
    if rp, exists := r.repo[tab]; exists {
        return rp, nil
    }
    return nil, fmt.Errorf("repository %s not configured", tab)
}

type ProviderRegistry struct {
    processors map[model.PaymentProvider]ports.IPaymentProcessor
}

func NewProviderRegistry() *ProviderRegistry {
    return &ProviderRegistry{
        processors: make(map[model.PaymentProvider]ports.IPaymentProcessor),
    }
}

func (r *ProviderRegistry) Register(provider model.PaymentProvider, processor ports.IPaymentProcessor) {
    r.processors[provider] = processor
}

func (r *ProviderRegistry) Get(provider model.PaymentProvider) (ports.IPaymentProcessor, error) {
    if p, exists := r.processors[provider]; exists {
        return p, nil
    }
    return nil, fmt.Errorf("provider %s not configured", provider)
}