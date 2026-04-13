CREATE TABLE IF NOT EXISTS invoice_settings (
    settings_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL UNIQUE,
    bankgiro VARCHAR(20),
    plusgiro VARCHAR(20),
    swish VARCHAR(20),
    iban VARCHAR(34),
    bic VARCHAR(11),
    f_skatt_text VARCHAR(500) DEFAULT 'Godkänd för F-skatt',
    default_payment_terms_days INT NOT NULL DEFAULT 30,
    next_invoice_number INT NOT NULL DEFAULT 1,
    invoice_prefix VARCHAR(10) DEFAULT '',
    default_revenue_account INT DEFAULT 3010,
    default_payment_account INT DEFAULT 1930,
    footer_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
);
