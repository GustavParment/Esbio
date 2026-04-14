CREATE TABLE IF NOT EXISTS invoice_lines (
    line_id SERIAL PRIMARY KEY,
    invoice_id INT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    description TEXT NOT NULL,
    quantity DECIMAL(15, 4) NOT NULL DEFAULT 1,
    unit VARCHAR(20) DEFAULT 'st',
    unit_price DECIMAL(15, 2) NOT NULL,
    discount_percent DECIMAL(5, 2) NOT NULL DEFAULT 0,
    vat_rate INT NOT NULL DEFAULT 25 CHECK (vat_rate IN (0, 6, 12, 25)),
    line_total DECIMAL(15, 2) NOT NULL,
    vat_amount DECIMAL(15, 2) NOT NULL,
    account_no INT NOT NULL DEFAULT 3010,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE,
    FOREIGN KEY (account_no) REFERENCES accounts(account_no) ON DELETE RESTRICT
);

CREATE INDEX idx_invoice_lines_invoice ON invoice_lines(invoice_id);
