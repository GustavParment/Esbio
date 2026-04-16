-- Bank feed tables (Tink Open Banking)

CREATE TABLE IF NOT EXISTS bank_connections (
    connection_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    created_by INT NOT NULL,
    tink_user_id VARCHAR(255) NOT NULL,
    tink_credentials_id VARCHAR(255),
    bank_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expiring_soon', 'expired', 'revoked')),
    access_token_encrypted TEXT,
    consent_expires_at TIMESTAMP NOT NULL,
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_bank_connections_company ON bank_connections(company_id);
CREATE INDEX IF NOT EXISTS idx_bank_connections_status ON bank_connections(status);

CREATE TABLE IF NOT EXISTS bank_accounts (
    bank_account_id SERIAL PRIMARY KEY,
    connection_id INT NOT NULL,
    company_id INT NOT NULL,
    tink_account_id VARCHAR(255) NOT NULL UNIQUE,
    account_name VARCHAR(255) NOT NULL,
    iban VARCHAR(50),
    account_number VARCHAR(50),
    currency VARCHAR(3) NOT NULL DEFAULT 'SEK',
    balance_amount DECIMAL(15, 2),
    balance_updated_at TIMESTAMP,
    mapped_account_no INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (connection_id) REFERENCES bank_connections(connection_id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (mapped_account_no) REFERENCES accounts(account_no) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_bank_accounts_connection ON bank_accounts(connection_id);
CREATE INDEX IF NOT EXISTS idx_bank_accounts_company ON bank_accounts(company_id);
CREATE INDEX IF NOT EXISTS idx_bank_accounts_tink_id ON bank_accounts(tink_account_id);

CREATE TABLE IF NOT EXISTS bank_transactions (
    bank_transaction_id SERIAL PRIMARY KEY,
    bank_account_id INT NOT NULL,
    company_id INT NOT NULL,
    tink_transaction_id VARCHAR(255) NOT NULL UNIQUE,
    booked_date DATE NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'SEK',
    description_original TEXT NOT NULL,
    description_display TEXT,
    merchant_name VARCHAR(255),
    merchant_category_code VARCHAR(10),
    tink_category_id VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'BOOKED' CHECK (status IN ('BOOKED', 'PENDING')),
    import_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (import_status IN ('pending', 'suggested', 'approved', 'skipped', 'booked')),
    suggested_account_no INT,
    suggested_description TEXT,
    voucher_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(bank_account_id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (suggested_account_no) REFERENCES accounts(account_no) ON DELETE SET NULL,
    FOREIGN KEY (voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_bank_txn_company ON bank_transactions(company_id);
CREATE INDEX IF NOT EXISTS idx_bank_txn_account ON bank_transactions(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_bank_txn_tink_id ON bank_transactions(tink_transaction_id);
CREATE INDEX IF NOT EXISTS idx_bank_txn_import_status ON bank_transactions(import_status);
CREATE INDEX IF NOT EXISTS idx_bank_txn_booked_date ON bank_transactions(booked_date);
CREATE INDEX IF NOT EXISTS idx_bank_txn_voucher ON bank_transactions(voucher_id);

CREATE TABLE IF NOT EXISTS bank_categorization_rules (
    rule_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    match_pattern VARCHAR(500) NOT NULL,
    match_type VARCHAR(20) NOT NULL DEFAULT 'contains' CHECK (match_type IN ('exact', 'contains', 'regex')),
    account_no INT NOT NULL,
    description_template TEXT,
    tax_code INT,
    confidence DECIMAL(3, 2) NOT NULL DEFAULT 1.00,
    use_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (account_no) REFERENCES accounts(account_no) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_categorization_rules_company ON bank_categorization_rules(company_id);
CREATE INDEX IF NOT EXISTS idx_categorization_rules_pattern ON bank_categorization_rules(match_pattern);
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'categorization_rules_company_pattern_unique') THEN
    ALTER TABLE bank_categorization_rules ADD CONSTRAINT categorization_rules_company_pattern_unique UNIQUE (company_id, match_pattern, match_type);
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS tink_oauth_state (
    state_id VARCHAR(64) PRIMARY KEY,
    company_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);
