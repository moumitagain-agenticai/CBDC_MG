-- Create ledger accounts table
CREATE TABLE IF NOT EXISTS cbdc_ledger_accounts (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL,
    category VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance DECIMAL(38,18) DEFAULT 0,
    debit_balance DECIMAL(38,18) DEFAULT 0,
    credit_balance DECIMAL(38,18) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    is_revaluation BOOLEAN DEFAULT FALSE,
    parent_id VARCHAR(36),
    tenant_id VARCHAR(36) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_accounts_code ON cbdc_ledger_accounts(code);
CREATE INDEX idx_ledger_accounts_type ON cbdc_ledger_accounts(type);
CREATE INDEX idx_ledger_accounts_category ON cbdc_ledger_accounts(category);
CREATE INDEX idx_ledger_accounts_tenant ON cbdc_ledger_accounts(tenant_id);

-- Create ledger entries table
CREATE TABLE IF NOT EXISTS cbdc_ledger_entries (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    account_code VARCHAR(50) NOT NULL,
    entry_type VARCHAR(10) NOT NULL,
    amount DECIMAL(38,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance_after DECIMAL(38,18) NOT NULL,
    description TEXT,
    reference_id VARCHAR(36),
    reference_type VARCHAR(50),
    tenant_id VARCHAR(36) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_transaction ON cbdc_ledger_entries(transaction_id);
CREATE INDEX idx_ledger_entries_account ON cbdc_ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_type ON cbdc_ledger_entries(entry_type);
CREATE INDEX idx_ledger_entries_tenant ON cbdc_ledger_entries(tenant_id);
CREATE INDEX idx_ledger_entries_created_at ON cbdc_ledger_entries(created_at);

-- Create currency positions table
CREATE TABLE IF NOT EXISTS cbdc_currency_positions (
    id VARCHAR(36) PRIMARY KEY,
    currency VARCHAR(3) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    long_position DECIMAL(38,18) DEFAULT 0,
    short_position DECIMAL(38,18) DEFAULT 0,
    net_position DECIMAL(38,18) DEFAULT 0,
    total_inflow DECIMAL(38,18) DEFAULT 0,
    total_outflow DECIMAL(38,18) DEFAULT 0,
    revaluation_rate DECIMAL(38,18) DEFAULT 1,
    revaluation_gain DECIMAL(38,18) DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_positions_currency ON cbdc_currency_positions(currency);
CREATE INDEX idx_positions_tenant ON cbdc_currency_positions(tenant_id);
CREATE INDEX idx_positions_status ON cbdc_currency_positions(status);

-- Create revaluations table
CREATE TABLE IF NOT EXISTS cbdc_revaluations (
    id VARCHAR(36) PRIMARY KEY,
    currency VARCHAR(3) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    old_rate DECIMAL(38,18) NOT NULL,
    new_rate DECIMAL(38,18) NOT NULL,
    old_position DECIMAL(38,18) NOT NULL,
    new_position DECIMAL(38,18) NOT NULL,
    gain_loss DECIMAL(38,18) NOT NULL,
    gain_loss_type VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL,
    revaluation_date TIMESTAMP WITH TIME ZONE NOT NULL,
    reference_id VARCHAR(36),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_revaluations_currency ON cbdc_revaluations(currency);
CREATE INDEX idx_revaluations_tenant ON cbdc_revaluations(tenant_id);
CREATE INDEX idx_revaluations_status ON cbdc_revaluations(status);
CREATE INDEX idx_revaluations_date ON cbdc_revaluations(revaluation_date);