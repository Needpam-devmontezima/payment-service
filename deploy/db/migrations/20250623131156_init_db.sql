-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- +goose StatementEnd

-- +goose StatementBegin
CREATE SEQUENCE payment_reference_seq START 1001;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    internal_reference VARCHAR(255) NOT NULL UNIQUE DEFAULT 'PAY-' || nextval('payment_reference_seq'),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency CHAR(3) NOT NULL,
    payment_intent_id VARCHAR(255) NOT NULL,
    tx_status VARCHAR(50) NOT NULL,
    customer_id VARCHAR(50) NOT NULL,
    save_payment_method BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payment_methods (
    customer_id VARCHAR(50) NOT NULL,
    payment_method_id VARCHAR(50) NOT NULL,
    pm_provider VARCHAR(50) NOT NULL CHECK (pm_provider IN ('stripe', 'paypal', 'needpam')),
    method_type VARCHAR(50) NULL DEFAULT 'none',
    pm_status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (pm_status IN ('active', 'disable')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (customer_id, payment_method_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    internal_reference VARCHAR(50) NOT NULL UNIQUE,
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    reason TEXT,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_customer_id ON transactions(customer_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_status ON transactions(tx_status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_payment_methods_customer ON payment_methods(customer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payment_methods_customer;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_customer_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS refunds;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS payment_methods;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd

-- +goose StatementBegin
DROP SEQUENCE IF EXISTS payment_reference_seq;
-- +goose StatementEnd
