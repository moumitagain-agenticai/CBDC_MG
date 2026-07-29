-- Create fees table
CREATE TABLE IF NOT EXISTS cbdc_fees (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    structure VARCHAR(20) NOT NULL,
    value DECIMAL(38,18) NOT NULL,
    min_amount DECIMAL(38,18),
    max_amount DECIMAL(38,18),
    tiered_structure JSONB,
    corridor_id VARCHAR(36),
    source_country VARCHAR(3),
    target_country VARCHAR(3),
    source_currency VARCHAR(3),
    target_currency VARCHAR(3),
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fees_code ON cbdc_fees(code);
CREATE INDEX idx_fees_type ON cbdc_fees(type);
CREATE INDEX idx_fees_corridor ON cbdc_fees(corridor_id);
CREATE INDEX idx_fees_source_country ON cbdc_fees(source_country);
CREATE INDEX idx_fees_target_country ON cbdc_fees(target_country);
CREATE INDEX idx_fees_is_active ON cbdc_fees(is_active);

-- Create fee corridors table
CREATE TABLE IF NOT EXISTS cbdc_fee_corridors (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    source_country VARCHAR(3) NOT NULL,
    target_country VARCHAR(3) NOT NULL,
    source_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    base_fee DECIMAL(38,18) NOT NULL,
    markup DECIMAL(38,18) DEFAULT 0,
    discount DECIMAL(38,18) DEFAULT 0,
    min_fee DECIMAL(38,18),
    max_fee DECIMAL(38,18),
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_corridors_code ON cbdc_fee_corridors(code);
CREATE INDEX idx_corridors_source_country ON cbdc_fee_corridors(source_country);
CREATE INDEX idx_corridors_target_country ON cbdc_fee_corridors(target_country);
CREATE INDEX idx_corridors_is_active ON cbdc_fee_corridors(is_active);

-- Create fee calculations table
CREATE TABLE IF NOT EXISTS cbdc_fee_calculations (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    amount DECIMAL(38,18) NOT NULL,
    total_fee DECIMAL(38,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    fee_breakdown JSONB NOT NULL,
    corridor_id VARCHAR(36),
    status VARCHAR(20) NOT NULL,
    metadata JSONB,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_calculations_transaction ON cbdc_fee_calculations(transaction_id);
CREATE INDEX idx_calculations_status ON cbdc_fee_calculations(status);
CREATE INDEX idx_calculations_created_at ON cbdc_fee_calculations(created_at);