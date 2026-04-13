CREATE TABLE IF NOT EXISTS invoices (
    invoice_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    customer_id INT NOT NULL,
    invoice_number INT NOT NULL,
    ocr_number VARCHAR(25) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'sent', 'paid', 'overdue', 'cancelled')),
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    payment_terms_days INT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'SEK',
    subtotal DECIMAL(15, 2) NOT NULL DEFAULT 0,
    vat_total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    rounding DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    amount_paid DECIMAL(15, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    your_reference VARCHAR(255),
    our_reference VARCHAR(255),
    bankgiro VARCHAR(20),
    plusgiro VARCHAR(20),
    f_skatt_text VARCHAR(500),
    sales_voucher_id INT,
    payment_voucher_id INT,
    payment_account INT DEFAULT 1930,
    revenue_account INT DEFAULT 3010,
    paid_at TIMESTAMP,
    sent_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE RESTRICT,
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id) ON DELETE RESTRICT,
    FOREIGN KEY (sales_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL,
    FOREIGN KEY (payment_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_invoices_company_number ON invoices(company_id, invoice_number);
CREATE INDEX idx_invoices_company ON invoices(company_id);
CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_status ON invoices(company_id, status);
CREATE INDEX idx_invoices_due_date ON invoices(company_id, due_date);
