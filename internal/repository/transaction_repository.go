package repository

import (
    "log"
	"fmt"
	"time"
    "errors"
    "strings"
	"context"
    
	"github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/danielmoisemontezima/zw-payment-service/internal/model"
)

// Combines all needed interfaces
type Queryable interface {
    Query(context.Context, string, ...any) (pgx.Rows, error)
    QueryRow(context.Context, string, ...any) pgx.Row
    Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
    CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type DB interface {
    Queryable
    Begin(ctx context.Context) (pgx.Tx, error)
}

type TransactionRepository struct {
    db DB
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: pool}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *model.Transaction) error {
    // Initialize builder
    var (
        fields []string
        values []interface{}
        params []string
        pos    = 1 // PostgreSQL parameter position counter
    )

    // Required fields (assuming these can't be null)
    fields = append(fields, "amount")
    values = append(values, tx.Amount)
    params = append(params, fmt.Sprintf("$%d", pos))
    pos++

    fields = append(fields, "currency")
    values = append(values, tx.Currency)
    params = append(params, fmt.Sprintf("$%d", pos))
    pos++

    fields = append(fields, "customer_id")
    values = append(values, tx.CustomerID)
    params = append(params, fmt.Sprintf("$%d", pos))
    pos++

    fields = append(fields, "tx_status")
    values = append(values, tx.TxStatus)
    params = append(params, fmt.Sprintf("$%d", pos))
    pos++

    // Conditionally add optional fields
    if tx.ID != "" {
        fields = append(fields, "id")
        values = append(values, tx.ID)
        params = append(params, fmt.Sprintf("$%d", pos))
        pos++
    }

    if tx.InternalReference != "" {
        fields = append(fields, "internal_reference")
        values = append(values, tx.InternalReference)
        params = append(params, fmt.Sprintf("$%d", pos))
        pos++
    }

    if tx.PaymentIntentID != "" {
        fields = append(fields, "payment_intent_id")
        values = append(values, tx.PaymentIntentID)
        params = append(params, fmt.Sprintf("$%d", pos))
        pos++
    }

    if tx.SavePaymentMethod != nil {
        fields = append(fields, "save_payment_method")
        values = append(values, *tx.SavePaymentMethod)
        params = append(params, fmt.Sprintf("$%d", pos))
        pos++
    }

    if tx.Metadata != nil {
        fields = append(fields, "metadata")
        values = append(values, tx.Metadata)
        params = append(params, fmt.Sprintf("$%d", pos))
        pos++
    }

    // Build the query
    query := fmt.Sprintf(`
        INSERT INTO transactions (%s)
        VALUES (%s)`,
        strings.Join(fields, ", "),
        strings.Join(params, ", "),
    )

    _, err := r.db.Exec(ctx, query, values...)
    return err
}

func (r *TransactionRepository) FindByID(ctx context.Context, id string) (*model.Transaction, error) {
	sql := `SELECT * FROM transactions WHERE id = $1`
	var tx model.Transaction
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&tx.ID,
		&tx.InternalReference,
		&tx.Amount,
		&tx.Currency,
		&tx.PaymentIntentID,
		&tx.TxStatus,
		&tx.CustomerID,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&tx.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction: %w", err)
	}
	return &tx, nil
}

func (r *TransactionRepository) FindByPaymentIntent(ctx context.Context, id string) (*model.Transaction, error) {
	sql := `SELECT id, internal_reference, amount, currency, payment_intent_id, tx_status, customer_id, save_payment_method, created_at, updated_at, metadata 
        FROM transactions WHERE payment_intent_id = $1`
	var tx model.Transaction
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&tx.ID,
		&tx.InternalReference,
		&tx.Amount,
		&tx.Currency,
		&tx.PaymentIntentID,
		&tx.TxStatus,
		&tx.CustomerID,
        &tx.SavePaymentMethod,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&tx.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction: %w", err)
	}
    log.Printf("Transaction-info-inside: %+v", &tx)
	return &tx, nil
}

func (r *TransactionRepository) FindByStatus(ctx context.Context, status string, since time.Time) ([]model.Transaction, error) {
	    const sql = `
        SELECT id, internal_reference, amount, currency, 
               payment_intent_id, status, customer_id,
               created_at, updated_at, metadata
        FROM transactions
        WHERE tx_status = $1 AND created_at >= $2
        ORDER BY created_at DESC
    `

    rows, err := r.db.Query(ctx, sql, status, since)
    if err != nil {
        return nil, fmt.Errorf("error querying transactions by status: %w", err)
    }
    defer rows.Close()

    var transactions []model.Transaction
    for rows.Next() {
        var tx model.Transaction
        if err := rows.Scan(
            &tx.ID,
            &tx.InternalReference,
            &tx.Amount,
            &tx.Currency,
            &tx.PaymentIntentID,
            &tx.TxStatus,
            &tx.CustomerID,
            &tx.CreatedAt,
            &tx.UpdatedAt,
            &tx.Metadata,
        ); err != nil {
            return nil, fmt.Errorf("error scanning transaction: %w", err)
        }
        transactions = append(transactions, tx)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows iteration error: %w", err)
    }

    return transactions, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, status string, filters map[string]interface{}) error {
    // Build dynamic SQL (pgx uses $1, $2... placeholders)
    var (
        query strings.Builder
        args  []interface{}
        pos   = 1 // Tracks placeholder position
    )

    // Base UPDATE
    query.WriteString(`UPDATE transactions SET tx_status = $1, updated_at = NOW()`)
    args = append(args, status)
    pos++

    // Add WHERE conditions if filters exist
    if len(filters) > 0 {
        query.WriteString(" WHERE ")
        conditions := []string{}
        
        for field, value := range filters {
            conditions = append(conditions, fmt.Sprintf("%s = $%d", field, pos))
            args = append(args, value)
            pos++
        }
        query.WriteString(strings.Join(conditions, " AND "))
    } else {
        return errors.New("at least one filter condition is required")
    }

    // Execute with pgx
    tag, err := r.db.Exec(ctx, query.String(), args...)
    if err != nil {
        return fmt.Errorf("failed to update transaction status: %w", err)
    }

    // Check if any rows were updated (pgx uses 'tag.RowsAffected()')
    if tag.RowsAffected() == 0 {
        return fmt.Errorf("no matching transaction found")
    }

    return nil
}

func (r *TransactionRepository) WithTransaction(ctx context.Context, 
    fn func(*TransactionRepository) error,
) error {
    // Begin a transaction (default options)
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // Create a transaction-scoped repository
    txRepo := &TransactionRepository{db: tx}

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback(ctx)
            panic(p) // Re-throw panic after cleanup
        }
    }()

    if err := fn(txRepo); err != nil {
        if rbErr := tx.Rollback(ctx); rbErr != nil {
            return fmt.Errorf("transaction error: %v, rollback failed: %w", err, rbErr)
        }
        return err
    }

    return tx.Commit(ctx)
}

func (r *TransactionRepository) FindByColumn(ctx context.Context, column string, value interface{}) ([]model.Transaction, error) {
    // Validate the column name to prevent SQL injection
    validColumns := map[string]bool{
        "id":                  true,
        "internal_reference":  true,
        "amount":             true,
        "currency":           true,
        "payment_intent_id":  true,
        "tx_status":           true,
        "customer_id":         true,
        "save_payment_method": true,
        "created_at":          true,
        "updated_at":          true,
        "metadata":            true,
    }

    if !validColumns[column] {
        return nil, fmt.Errorf("invalid column name: %s", column)
    }

    sql := fmt.Sprintf(`
        SELECT id, internal_reference, amount, currency, 
               payment_intent_id, tx_status, customer_id,
               save_payment_method, created_at, updated_at, metadata
        FROM transactions
        WHERE %s = $1
        ORDER BY created_at DESC`, column)

    rows, err := r.db.Query(ctx, sql, value)
    if err != nil {
        return nil, fmt.Errorf("error querying transactions by %s: %w", column, err)
    }
    defer rows.Close()

    var transactions []model.Transaction
    for rows.Next() {
        var tx model.Transaction
        if err := rows.Scan(
            &tx.ID,
            &tx.InternalReference,
            &tx.Amount,
            &tx.Currency,
            &tx.PaymentIntentID,
            &tx.TxStatus,
            &tx.CustomerID,
            &tx.SavePaymentMethod,
            &tx.CreatedAt,
            &tx.UpdatedAt,
            &tx.Metadata,
        ); err != nil {
            return nil, fmt.Errorf("error scanning transaction: %w", err)
        }
        transactions = append(transactions, tx)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows iteration error: %w", err)
    }

    return transactions, nil
}