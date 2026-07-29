-- Create transactions table
CREATE TABLE IF NOT EXISTS cbdc_transactions (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    state VARCHAR(30) NOT NULL,
    status VARCHAR(30) NOT NULL,
    source_country VARCHAR(3) NOT NULL,
    target_country VARCHAR(3) NOT NULL,
    source_account_id VARCHAR(100) NOT NULL,
    target_account_id VARCHAR(100),
    source_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    source_amount DECIMAL(38,18) NOT NULL,
    target_amount DECIMAL(38,18),
    conversion_rate DECIMAL(38,18),
    lock_reference VARCHAR(100),
    settlement_id VARCHAR(36),
    error_message TEXT,
    cancel_reason TEXT,
    idempotency_key VARCHAR(100) UNIQUE,
    attempts INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    version BIGINT DEFAULT 0
);

-- Create indexes for performance
CREATE INDEX idx_cbdc_transactions_status ON cbdc_transactions(status);
CREATE INDEX idx_cbdc_transactions_state ON cbdc_transactions(state);
CREATE INDEX idx_cbdc_transactions_idempotency_key ON cbdc_transactions(idempotency_key);
CREATE INDEX idx_cbdc_transactions_source_country ON cbdc_transactions(source_country);
CREATE INDEX idx_cbdc_transactions_target_country ON cbdc_transactions(target_country);
CREATE INDEX idx_cbdc_transactions_created_at ON cbdc_transactions(created_at);
CREATE INDEX idx_cbdc_transactions_updated_at ON cbdc_transactions(updated_at);