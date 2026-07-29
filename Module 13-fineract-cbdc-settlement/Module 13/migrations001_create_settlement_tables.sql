-- Create settlements table
CREATE TABLE IF NOT EXISTS cbdc_settlements (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    source_network VARCHAR(20) NOT NULL,
    target_network VARCHAR(20) NOT NULL,
    source_account_id VARCHAR(100) NOT NULL,
    target_account_id VARCHAR(100) NOT NULL,
    source_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    source_amount DECIMAL(38,18) NOT NULL,
    target_amount DECIMAL(38,18) NOT NULL,
    conversion_rate DECIMAL(38,18),
    source_lock_id VARCHAR(100),
    target_lock_id VARCHAR(100),
    burn_transaction_id VARCHAR(100),
    issue_transaction_id VARCHAR(100),
    status VARCHAR(30) NOT NULL,
    type VARCHAR(20) NOT NULL,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    rolled_back_at TIMESTAMP WITH TIME ZONE,
    version BIGINT DEFAULT 0
);

CREATE INDEX idx_settlements_transaction ON cbdc_settlements(transaction_id);
CREATE INDEX idx_settlements_status ON cbdc_settlements(status);
CREATE INDEX idx_settlements_source_network ON cbdc_settlements(source_network);
CREATE INDEX idx_settlements_target_network ON cbdc_settlements(target_network);
CREATE INDEX idx_settlements_created_at ON cbdc_settlements(created_at);

-- Create fund locks table
CREATE TABLE IF NOT EXISTS cbdc_fund_locks (
    id VARCHAR(36) PRIMARY KEY,
    settlement_id VARCHAR(36) NOT NULL,
    network VARCHAR(20) NOT NULL,
    account_id VARCHAR(100) NOT NULL,
    lock_id VARCHAR(100) UNIQUE NOT NULL,
    amount DECIMAL(38,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL,
    lock_duration BIGINT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    released_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB
);

CREATE INDEX idx_locks_settlement ON cbdc_fund_locks(settlement_id);
CREATE INDEX idx_locks_lock_id ON cbdc_fund_locks(lock_id);
CREATE INDEX idx_locks_status ON cbdc_fund_locks(status);

-- Create burn records table
CREATE TABLE IF NOT EXISTS cbdc_burn_records (
    id VARCHAR(36) PRIMARY KEY,
    settlement_id VARCHAR(36) NOT NULL,
    network VARCHAR(20) NOT NULL,
    lock_id VARCHAR(100) NOT NULL,
    transaction_id VARCHAR(100) UNIQUE NOT NULL,
    amount DECIMAL(38,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL,
    confirmation_block VARCHAR(100),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB
);

CREATE INDEX idx_burns_settlement ON cbdc_burn_records(settlement_id);
CREATE INDEX idx_burns_transaction ON cbdc_burn_records(transaction_id);
CREATE INDEX idx_burns_status ON cbdc_burn_records(status);

-- Create compensations table
CREATE TABLE IF NOT EXISTS cbdc_compensations (
    id VARCHAR(36) PRIMARY KEY,
    settlement_id VARCHAR(36) NOT NULL,
    original_burn_tx VARCHAR(100) NOT NULL,
    network VARCHAR(20) NOT NULL,
    account_id VARCHAR(100) NOT NULL,
    amount DECIMAL(38,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL,
    compensation_tx VARCHAR(100),
    reason TEXT,
    alert_sent BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB
);

CREATE INDEX idx_compensations_settlement ON cbdc_compensations(settlement_id);
CREATE INDEX idx_compensations_burn_tx ON cbdc_compensations(original_burn_tx);
CREATE INDEX idx_compensations_status ON cbdc_compensations(status);