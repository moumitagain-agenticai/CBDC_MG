-- Create screenings table
CREATE TABLE IF NOT EXISTS cbdc_screenings (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    customer_id VARCHAR(36) NOT NULL,
    customer_name VARCHAR(200) NOT NULL,
    customer_country VARCHAR(3) NOT NULL,
    amount DECIMAL(38,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    source_country VARCHAR(3) NOT NULL,
    target_country VARCHAR(3) NOT NULL,
    status VARCHAR(30) NOT NULL,
    type VARCHAR(20) NOT NULL,
    result TEXT,
    score INTEGER DEFAULT 0,
    details JSONB,
    matched_sanctions JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_screenings_transaction ON cbdc_screenings(transaction_id);
CREATE INDEX idx_screenings_customer ON cbdc_screenings(customer_id);
CREATE INDEX idx_screenings_status ON cbdc_screenings(status);
CREATE INDEX idx_screenings_type ON cbdc_screenings(type);
CREATE INDEX idx_screenings_created_at ON cbdc_screenings(created_at);

-- Create sanctions lists table
CREATE TABLE IF NOT EXISTS cbdc_sanctions_lists (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    aliases JSONB,
    type VARCHAR(50) NOT NULL,
    source VARCHAR(50) NOT NULL,
    country VARCHAR(3),
    nationality VARCHAR(3),
    date_of_birth VARCHAR(20),
    identification JSONB,
    reasons JSONB,
    listed_date TIMESTAMP WITH TIME ZONE NOT NULL,
    unlisted_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sanctions_name ON cbdc_sanctions_lists(name);
CREATE INDEX idx_sanctions_source ON cbdc_sanctions_lists(source);
CREATE INDEX idx_sanctions_type ON cbdc_sanctions_lists(type);
CREATE INDEX idx_sanctions_is_active ON cbdc_sanctions_lists(is_active);

-- Create compliance checks table
CREATE TABLE IF NOT EXISTS cbdc_compliance_checks (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    screening_id VARCHAR(36),
    customer_id VARCHAR(36) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(30) NOT NULL,
    result JSONB,
    rules_applied JSONB,
    score INTEGER DEFAULT 0,
    details JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_compliance_transaction ON cbdc_compliance_checks(transaction_id);
CREATE INDEX idx_compliance_screening ON cbdc_compliance_checks(screening_id);
CREATE INDEX idx_compliance_customer ON cbdc_compliance_checks(customer_id);
CREATE INDEX idx_compliance_status ON cbdc_compliance_checks(status);

-- Create audit trail table
CREATE TABLE IF NOT EXISTS cbdc_audit_trails (
    id VARCHAR(36) PRIMARY KEY,
    screening_id VARCHAR(36),
    action VARCHAR(50) NOT NULL,
    status VARCHAR(30) NOT NULL,
    user_id VARCHAR(36),
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_screening ON cbdc_audit_trails(screening_id);
CREATE INDEX idx_audit_action ON cbdc_audit_trails(action);
CREATE INDEX idx_audit_user ON cbdc_audit_trails(user_id);
CREATE INDEX idx_audit_created_at ON cbdc_audit_trails(created_at);