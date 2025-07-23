package repository

import (
	"fmt"
	"time"
    "errors"
    "strings"
	"context"
    "database/sql"
    
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/danielmoisemontezima/zw-payment-service/internal/model"
)

type PaymentMethodRepository struct {
    db DB
}

func NewPaymentMethodRepository(pool *pgxpool.Pool) *PaymentMethodRepository {
	return &PaymentMethodRepository{db: pool}
}

func (r *PaymentMethodRepository) Create(ctx context.Context, tx *model.Transaction) error {
    // Initialize builder
    var (
        fields []string
        values []interface{}
        params []string
        pos    = 1 // PostgreSQL parameter position counter
    )

    // Required fields (assuming these can't be null)
    fields = append(fields, "customer_id")
    values = append(values, tx.CustomerID)
    params = append(params, fmt.Sprintf("$%d", pos))
    pos++

    fields = append(fields, "payment_method_id")
    values = append(values, tx.PaymentMethodID)
    params = append(params, fmt.Sprintf("$%d", pos))
    pos++

    fields = append(fields, "pm_provider")
    values = append(values, tx.PaymentProvider)
    params = append(params, fmt.Sprintf("$%d", pos))
    pos++

    // Conditionally add optional fields
    if tx.PaymentMethodType != "" {
        fields = append(fields, "method_type")
        values = append(values, tx.PaymentMethodType)
        params = append(params, fmt.Sprintf("$%d", pos))
        pos++
    }

    if tx.PaymentMethodStatus != "" {
        fields = append(fields, "pm_status")
        values = append(values, tx.PaymentMethodStatus)
        params = append(params, fmt.Sprintf("$%d", pos))
        pos++
    }

    // Build the query
    query := fmt.Sprintf(`
        INSERT INTO payment_methods (%s)
        VALUES (%s)`,
        strings.Join(fields, ", "),
        strings.Join(params, ", "),
    )

    _, err := r.db.Exec(ctx, query, values...)
    return err
}

func (r *PaymentMethodRepository) FindByID(ctx context.Context, id string) (*model.Transaction, error) {
	rawsql := `SELECT customer_id, payment_method_id, pm_provider, method_type, pm_status
               FROM payment_methods WHERE payment_method_id = $1`
               
	var tx model.Transaction
	err := r.db.QueryRow(ctx, rawsql, id).Scan(
		&tx.CustomerID,
		&tx.PaymentMethodID,
		&tx.PaymentProvider,
		&tx.PaymentMethodType,
		&tx.PaymentMethodStatus,
	)
	if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("Not found: %w", err)
        }
        return nil, fmt.Errorf("error getting payment Method: %w", err)
	}
	return &tx, nil
}

func (r *PaymentMethodRepository) FindByPaymentIntent(ctx context.Context, id string) (*model.Transaction, error) {
	var tx model.Transaction
	return &tx, nil
}

func (r *PaymentMethodRepository) FindByStatus(ctx context.Context, status string, since time.Time) ([]model.Transaction, error) {
	    const rawsql = `
        SELECT customer_id, payment_method_id, pm_provider, method_type, 
        	   pm_status, created_at, updated_at
        FROM payment_methods
        WHERE pm_status = $1 AND created_at >= $2
        ORDER BY created_at DESC
    `

    rows, err := r.db.Query(ctx, rawsql, status, since)
    if err != nil {
        return nil, fmt.Errorf("error querying payment_methods by status: %w", err)
    }
    defer rows.Close()

    var payment_methods []model.Transaction
    for rows.Next() {
        var tx model.Transaction
        if err := rows.Scan(
			&tx.CustomerID,
			&tx.PaymentMethodID,
			&tx.PaymentProvider,
			&tx.PaymentMethodType,
			&tx.PaymentMethodStatus,
        ); err != nil {
            return nil, fmt.Errorf("error scanning payment method: %w", err)
        }
        payment_methods = append(payment_methods, tx)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows iteration error: %w", err)
    }

    return payment_methods, nil
}

func (r *PaymentMethodRepository) UpdateStatus(ctx context.Context, status string, filters map[string]interface{}) error {
    // Build dynamic SQL (pgx uses $1, $2... placeholders)
    var (
        query strings.Builder
        args  []interface{}
        pos   = 1 // Tracks placeholder position
    )

    // Base UPDATE
    query.WriteString(`UPDATE payment_methods SET pm_status = $1, updated_at = NOW()`)
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
        return fmt.Errorf("failed to update payment method status: %w", err)
    }

    // Check if any rows were updated (pgx uses 'tag.RowsAffected()')
    if tag.RowsAffected() == 0 {
        return fmt.Errorf("no matching payment method found")
    }

    return nil
}

func (r *PaymentMethodRepository) WithTransaction(ctx context.Context, 
    fn func(*PaymentMethodRepository) error,
) error {
    // Begin a transaction (default options)
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // Create a transaction-scoped repository
    txRepo := &PaymentMethodRepository{db: tx}

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

func (r *PaymentMethodRepository) FindByColumn(ctx context.Context, column string, value interface{}) ([]model.Transaction, error) {
    // Validate the column name to prevent SQL injection
    validColumns := map[string]bool{
        "customer_id":          true,
        "payment_method_id":    true,
        "pm_provider":          true,
        "method_type":          true,
        "pm_status":            true,
        "created_at":           true,
        "updated_at":           true,
    }

    if !validColumns[column] {
        return nil, fmt.Errorf("invalid column name: %s", column)
    }

    sql := fmt.Sprintf(`
        SELECT customer_id, payment_method_id, pm_provider, method_type, 
               pm_status
        FROM payment_methods
        WHERE %s = $1
        ORDER BY created_at DESC`, column)

    rows, err := r.db.Query(ctx, sql, value)
    if err != nil {
        return nil, fmt.Errorf("error querying payment method by %s: %w", column, err)
    }
    defer rows.Close()

    var payment_methods []model.Transaction
    for rows.Next() {
        var tx model.Transaction
        if err := rows.Scan(
			&tx.CustomerID,
			&tx.PaymentMethodID,
			&tx.PaymentProvider,
			&tx.PaymentMethodType,
			&tx.PaymentMethodStatus,
        ); err != nil {
            return nil, fmt.Errorf("error scanning payment method: %w", err)
        }
        payment_methods = append(payment_methods, tx)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows iteration error: %w", err)
    }

    return payment_methods, nil
}