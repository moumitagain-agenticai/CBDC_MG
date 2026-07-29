-- Create reconciliations table
CREATE TABLE IF NOT EXISTS cbdc_reconciliations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(30) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    opening_balance DECIMAL(38,18) NOT NULL,
    closing_balance DECIMAL(38,18) DEFAULT 0,
    system_balance DECIMAL(38,18) DEFAULT 0,
    bank_balance DECIMAL(38,18) DEFAULT 0,
    difference DECIMAL(38,18) DEFAULT 0,
    total_entries INTEGER DEFAULT 0,
    matched_entries INTEGER DEFAULT 0,
    unmatched_entries INTEGER DEFAULT 0,
    tenant_id VARCHAR(36) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_reconciliations_account ON cbdc_reconciliations(account_id);
CREATE INDEX idx_reconciliations_status ON cbdc_reconciliations(status);
CREATE INDEX idx_reconciliations_type ON cbdc_reconciliations(type);
CREATE INDEX idx_reconciliations_tenant ON cbdc_reconciliations(tenant_id);

-- Create bank statements table
CREATE TABLE IF NOT EXISTS cbdc_bank_statements (
    id VARCHAR(36) PRIMARY KEY,
    reconciliation_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    statement_date TIMESTAMP WITH TIME ZONE NOT NULL,
    statement_type VARCHAR(20) NOT NULL,
    opening_balance DECIMAL(38,18) NOT NULL,
    closing_balance DECIMAL(38,18) NOT NULL,
    total_debit DECIMAL(38,18) NOT NULL,
    total_credit DECIMAL(38,18) NOT NULL,
    entries JSONB NOT NULL,
    status VARCHAR(20) NOT NULL,
    file_name VARCHAR(200),
    file_content BYTEA,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_statements_reconciliation ON cbdc_bank_statements(reconciliation_id);
CREATE INDEX idx_statements_account ON cbdc_bank_statements(account_id);
CREATE INDEX idx_statements_status ON cbdc_bank_statements(status);

-- Create exception items table
CREATE TABLE IF NOT EXISTS cbdc_exception_items (
    id VARCHAR(36) PRIMARY KEY,
    reconciliation_id VARCHAR(36) NOT NULL,
    type VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL,
    priority VARCHAR(10) NOT NULL,
    description TEXT,
    amount DECIMAL(38,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    system_transaction_id VARCHAR(36),
    bank_transaction_id VARCHAR(100),
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    resolution TEXT,
    resolved_by VARCHAR(100),
    resolved_at TIMESTAMP WITH TIME ZONE,
    tenant_id VARCHAR(36) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exceptions_reconciliation ON cbdc_exception_items(reconciliation_id);
CREATE INDEX idx_exceptions_type ON cbdc_exception_items(type);
CREATE INDEX idx_exceptions_status ON cbdc_exception_items(status);
CREATE INDEX idx_exceptions_tenant ON cbdc_exception_items(tenant_id);