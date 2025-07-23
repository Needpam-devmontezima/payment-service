package ports

import (
	"time"
	"context"
    
	"github.com/danielmoisemontezima/zw-payment-service/internal/model"
)

type ITransactionRepository interface {
    Create(ctx context.Context, tx *model.Transaction) error
    FindByID(ctx context.Context, id string) (*model.Transaction, error)
    FindByPaymentIntent(ctx context.Context, id string) (*model.Transaction, error)
    FindByStatus(ctx context.Context, status string, since time.Time) ([]model.Transaction, error)
    UpdateStatus(ctx context.Context, status string, filters map[string]interface{}) error
    FindByColumn(ctx context.Context, column string, value interface{}) ([]model.Transaction, error)
}