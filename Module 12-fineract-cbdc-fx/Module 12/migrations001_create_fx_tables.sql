-- Create exchange rates table
CREATE TABLE IF NOT EXISTS cbdc_exchange_rates (
    id VARCHAR(36) PRIMARY KEY,
    base_currency VARCHAR(3) NOT NULL,
    quote_currency VARCHAR(3) NOT NULL,
    bid_rate DECIMAL(38,18) NOT NULL,
    ask_rate DECIMAL(38,18) NOT NULL,
    mid_rate DECIMAL(38,18) NOT NULL,
    spread DECIMAL(38,18) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exchange_rates_currencies ON cbdc_exchange_rates(base_currency, quote_currency);
CREATE INDEX idx_exchange_rates_status ON cbdc_exchange_rates(status);
CREATE INDEX idx_exchange_rates_expires_at ON cbdc_exchange_rates(expires_at);

-- Create FX quotes table
CREATE TABLE IF NOT EXISTS cbdc_fx_quotes (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    base_currency VARCHAR(3) NOT NULL,
    quote_currency VARCHAR(3) NOT NULL,
    base_amount DECIMAL(38,18) NOT NULL,
    quote_amount DECIMAL(38,18) NOT NULL,
    rate DECIMAL(38,18) NOT NULL,
    bid_rate DECIMAL(38,18) NOT NULL,
    ask_rate DECIMAL(38,18) NOT NULL,
    spread DECIMAL(38,18) NOT NULL,
    markup_percent DECIMAL(38,18) NOT NULL,
    markup_amount DECIMAL(38,18) NOT NULL,
    slippage_percent DECIMAL(38,18) DEFAULT 0,
    slippage_amount DECIMAL(38,18) DEFAULT 0,
    final_rate DECIMAL(38,18) NOT NULL,
    status VARCHAR(20) NOT NULL,
    lock_duration BIGINT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    locked_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fx_quotes_transaction ON cbdc_fx_quotes(transaction_id);
CREATE INDEX idx_fx_quotes_status ON cbdc_fx_quotes(status);
CREATE INDEX idx_fx_quotes_expires_at ON cbdc_fx_quotes(expires_at);

-- Create conversions table
CREATE TABLE IF NOT EXISTS cbdc_conversions (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    quote_id VARCHAR(36),
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    from_amount DECIMAL(38,18) NOT NULL,
    to_amount DECIMAL(38,18) NOT NULL,
    rate_used DECIMAL(38,18) NOT NULL,
    fee_amount DECIMAL(38,18) NOT NULL,
    fee_currency VARCHAR(3) NOT NULL,
    markup_applied DECIMAL(38,18) NOT NULL,
    slippage_applied DECIMAL(38,18) NOT NULL,
    status VARCHAR(20) NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB
);

CREATE INDEX idx_conversions_transaction ON cbdc_conversions(transaction_id);
CREATE INDEX idx_conversions_quote ON cbdc_conversions(quote_id);
CREATE INDEX idx_conversions_status ON cbdc_conversions(status);