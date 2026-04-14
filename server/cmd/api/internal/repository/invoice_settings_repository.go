package repository

import (
	"database/sql"
	"esbio/api/internal/domain"
	"fmt"
)

type InvoiceSettingsRepository interface {
	GetByCompanyID(companyID int) (*domain.InvoiceSettings, error)
	Upsert(settings *domain.InvoiceSettings) error
	GetAndIncrementInvoiceNumber(companyID int) (int, error)
}

type invoiceSettingsRepository struct {
	db *sql.DB
}

func NewInvoiceSettingsRepository(db *sql.DB) InvoiceSettingsRepository {
	return &invoiceSettingsRepository{db: db}
}

func (r *invoiceSettingsRepository) GetByCompanyID(companyID int) (*domain.InvoiceSettings, error) {
	query := `
		SELECT settings_id, company_id, COALESCE(bankgiro, ''), COALESCE(plusgiro, ''),
			COALESCE(swish, ''), COALESCE(iban, ''), COALESCE(bic, ''),
			COALESCE(f_skatt_text, 'Godkänd för F-skatt'), default_payment_terms_days,
			next_invoice_number, COALESCE(invoice_prefix, ''), default_revenue_account,
			default_payment_account, COALESCE(footer_text, ''), created_at, updated_at
		FROM invoice_settings WHERE company_id = $1
	`
	s := &domain.InvoiceSettings{}
	err := r.db.QueryRow(query, companyID).Scan(
		&s.SettingsID, &s.CompanyID, &s.Bankgiro, &s.Plusgiro,
		&s.Swish, &s.IBAN, &s.BIC, &s.FSkattText, &s.DefaultPaymentTermsDays,
		&s.NextInvoiceNumber, &s.InvoicePrefix, &s.DefaultRevenueAccount,
		&s.DefaultPaymentAccount, &s.FooterText, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Auto-create default settings for this company
		defaults := &domain.InvoiceSettings{
			CompanyID:               companyID,
			FSkattText:              "Godkänd för F-skatt",
			DefaultPaymentTermsDays: 30,
			NextInvoiceNumber:       1,
			DefaultRevenueAccount:   3010,
			DefaultPaymentAccount:   1930,
		}
		if err := r.Upsert(defaults); err != nil {
			return nil, fmt.Errorf("failed to create default settings: %w", err)
		}
		return defaults, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice settings: %w", err)
	}
	return s, nil
}

func (r *invoiceSettingsRepository) Upsert(settings *domain.InvoiceSettings) error {
	query := `
		INSERT INTO invoice_settings (
			company_id, bankgiro, plusgiro, swish, iban, bic, f_skatt_text,
			default_payment_terms_days, next_invoice_number, invoice_prefix,
			default_revenue_account, default_payment_account, footer_text
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (company_id) DO UPDATE SET
			bankgiro = EXCLUDED.bankgiro,
			plusgiro = EXCLUDED.plusgiro,
			swish = EXCLUDED.swish,
			iban = EXCLUDED.iban,
			bic = EXCLUDED.bic,
			f_skatt_text = EXCLUDED.f_skatt_text,
			default_payment_terms_days = EXCLUDED.default_payment_terms_days,
			invoice_prefix = EXCLUDED.invoice_prefix,
			default_revenue_account = EXCLUDED.default_revenue_account,
			default_payment_account = EXCLUDED.default_payment_account,
			footer_text = EXCLUDED.footer_text,
			updated_at = NOW()
		RETURNING settings_id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		settings.CompanyID, settings.Bankgiro, settings.Plusgiro, settings.Swish,
		settings.IBAN, settings.BIC, settings.FSkattText, settings.DefaultPaymentTermsDays,
		settings.NextInvoiceNumber, settings.InvoicePrefix, settings.DefaultRevenueAccount,
		settings.DefaultPaymentAccount, settings.FooterText,
	).Scan(&settings.SettingsID, &settings.CreatedAt, &settings.UpdatedAt)
}

// GetAndIncrementInvoiceNumber atomically assigns the next invoice number.
// Returns the assigned number (before increment).
func (r *invoiceSettingsRepository) GetAndIncrementInvoiceNumber(companyID int) (int, error) {
	// Ensure settings row exists
	_, err := r.GetByCompanyID(companyID)
	if err != nil {
		return 0, err
	}

	var num int
	err = r.db.QueryRow(`
		UPDATE invoice_settings
		SET next_invoice_number = next_invoice_number + 1, updated_at = NOW()
		WHERE company_id = $1
		RETURNING next_invoice_number - 1
	`, companyID).Scan(&num)
	if err != nil {
		return 0, fmt.Errorf("failed to increment invoice number: %w", err)
	}
	return num, nil
}
