package repository

import (
	"database/sql"
	"esbio/api/internal/domain"
	"fmt"

	"github.com/shopspring/decimal"
)

type InvoiceRepository interface {
	CreateInvoice(invoice *domain.Invoice) error
	GetInvoiceByID(invoiceID int) (*domain.Invoice, error)
	GetInvoicesByCompanyID(companyID int) ([]*domain.Invoice, error)
	GetInvoicesByStatus(companyID int, status string) ([]*domain.Invoice, error)
	GetInvoicesByCustomerID(companyID int, customerID int) ([]*domain.Invoice, error)
	GetOverdueInvoices(companyID int) ([]*domain.Invoice, error)
	UpdateInvoice(invoice *domain.Invoice) error
	UpdateInvoiceStatus(invoiceID int, status string) error
	LinkSalesVoucher(invoiceID int, voucherID int) error
	LinkPaymentVoucher(invoiceID int, voucherID int) error
	MarkAsPaid(invoiceID int, amountPaid decimal.Decimal, paidAt string) error
	MarkAsSent(invoiceID int, sentAt string) error
	MarkAsCancelled(invoiceID int, cancelledAt string) error
	DeleteInvoice(invoiceID int) error
}

type invoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) CreateInvoice(invoice *domain.Invoice) error {
	query := `
		INSERT INTO invoices (
			company_id, customer_id, invoice_number, ocr_number, status,
			invoice_date, due_date, payment_terms_days, currency,
			subtotal, vat_total, rounding, total,
			notes, your_reference, our_reference,
			bankgiro, plusgiro, f_skatt_text,
			payment_account, revenue_account, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING invoice_id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		invoice.CompanyID, invoice.CustomerID, invoice.InvoiceNumber, invoice.OCRNumber, invoice.Status,
		invoice.InvoiceDate.Time, invoice.DueDate.Time, invoice.PaymentTermsDays, invoice.Currency,
		invoice.Subtotal, invoice.VATTotal, invoice.Rounding, invoice.Total,
		invoice.Notes, invoice.YourReference, invoice.OurReference,
		invoice.Bankgiro, invoice.Plusgiro, invoice.FSkattText,
		invoice.PaymentAccount, invoice.RevenueAccount, invoice.CreatedBy,
	).Scan(&invoice.InvoiceID, &invoice.CreatedAt, &invoice.UpdatedAt)
}

func (r *invoiceRepository) GetInvoiceByID(invoiceID int) (*domain.Invoice, error) {
	query := `
		SELECT i.invoice_id, i.company_id, i.customer_id, i.invoice_number, i.ocr_number, i.status,
			i.invoice_date, i.due_date, i.payment_terms_days, i.currency,
			i.subtotal, i.vat_total, i.rounding, i.total, i.amount_paid,
			COALESCE(i.notes,''), COALESCE(i.your_reference,''), COALESCE(i.our_reference,''),
			COALESCE(i.bankgiro,''), COALESCE(i.plusgiro,''), COALESCE(i.f_skatt_text,''),
			i.sales_voucher_id, i.payment_voucher_id, i.payment_account, i.revenue_account,
			i.paid_at, i.sent_at, i.cancelled_at, i.created_by, i.created_at, i.updated_at
		FROM invoices i WHERE i.invoice_id = $1
	`
	inv := &domain.Invoice{}
	var paidAt, sentAt, cancelledAt sql.NullTime
	var salesVoucherID, paymentVoucherID sql.NullInt64
	err := r.db.QueryRow(query, invoiceID).Scan(
		&inv.InvoiceID, &inv.CompanyID, &inv.CustomerID, &inv.InvoiceNumber, &inv.OCRNumber, &inv.Status,
		&inv.InvoiceDate.Time, &inv.DueDate.Time, &inv.PaymentTermsDays, &inv.Currency,
		&inv.Subtotal, &inv.VATTotal, &inv.Rounding, &inv.Total, &inv.AmountPaid,
		&inv.Notes, &inv.YourReference, &inv.OurReference,
		&inv.Bankgiro, &inv.Plusgiro, &inv.FSkattText,
		&salesVoucherID, &paymentVoucherID, &inv.PaymentAccount, &inv.RevenueAccount,
		&paidAt, &sentAt, &cancelledAt, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if salesVoucherID.Valid {
		v := int(salesVoucherID.Int64)
		inv.SalesVoucherID = &v
	}
	if paymentVoucherID.Valid {
		v := int(paymentVoucherID.Int64)
		inv.PaymentVoucherID = &v
	}
	if paidAt.Valid {
		s := paidAt.Time.Format("2006-01-02T15:04:05Z")
		inv.PaidAt = &s
	}
	if sentAt.Valid {
		s := sentAt.Time.Format("2006-01-02T15:04:05Z")
		inv.SentAt = &s
	}
	if cancelledAt.Valid {
		s := cancelledAt.Time.Format("2006-01-02T15:04:05Z")
		inv.CancelledAt = &s
	}
	return inv, nil
}

func (r *invoiceRepository) GetInvoicesByCompanyID(companyID int) ([]*domain.Invoice, error) {
	return r.queryInvoices(`
		SELECT invoice_id, company_id, customer_id, invoice_number, ocr_number, status,
			invoice_date, due_date, payment_terms_days, currency,
			subtotal, vat_total, rounding, total, amount_paid, created_by, created_at, updated_at
		FROM invoices WHERE company_id = $1 ORDER BY invoice_number DESC
	`, companyID)
}

func (r *invoiceRepository) GetInvoicesByStatus(companyID int, status string) ([]*domain.Invoice, error) {
	return r.queryInvoices(`
		SELECT invoice_id, company_id, customer_id, invoice_number, ocr_number, status,
			invoice_date, due_date, payment_terms_days, currency,
			subtotal, vat_total, rounding, total, amount_paid, created_by, created_at, updated_at
		FROM invoices WHERE company_id = $1 AND status = $2 ORDER BY invoice_number DESC
	`, companyID, status)
}

func (r *invoiceRepository) GetInvoicesByCustomerID(companyID int, customerID int) ([]*domain.Invoice, error) {
	return r.queryInvoices(`
		SELECT invoice_id, company_id, customer_id, invoice_number, ocr_number, status,
			invoice_date, due_date, payment_terms_days, currency,
			subtotal, vat_total, rounding, total, amount_paid, created_by, created_at, updated_at
		FROM invoices WHERE company_id = $1 AND customer_id = $2 ORDER BY invoice_number DESC
	`, companyID, customerID)
}

func (r *invoiceRepository) GetOverdueInvoices(companyID int) ([]*domain.Invoice, error) {
	return r.queryInvoices(`
		SELECT invoice_id, company_id, customer_id, invoice_number, ocr_number, status,
			invoice_date, due_date, payment_terms_days, currency,
			subtotal, vat_total, rounding, total, amount_paid, created_by, created_at, updated_at
		FROM invoices WHERE company_id = $1 AND status = 'sent' AND due_date < CURRENT_DATE
		ORDER BY due_date
	`, companyID)
}

// queryInvoices is a helper for list queries (lighter scan, no detail fields)
func (r *invoiceRepository) queryInvoices(query string, args ...interface{}) ([]*domain.Invoice, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		inv := &domain.Invoice{}
		if err := rows.Scan(
			&inv.InvoiceID, &inv.CompanyID, &inv.CustomerID, &inv.InvoiceNumber, &inv.OCRNumber, &inv.Status,
			&inv.InvoiceDate.Time, &inv.DueDate.Time, &inv.PaymentTermsDays, &inv.Currency,
			&inv.Subtotal, &inv.VATTotal, &inv.Rounding, &inv.Total, &inv.AmountPaid,
			&inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}
	return invoices, nil
}

func (r *invoiceRepository) UpdateInvoice(invoice *domain.Invoice) error {
	query := `
		UPDATE invoices SET
			customer_id=$1, invoice_date=$2, due_date=$3, payment_terms_days=$4,
			subtotal=$5, vat_total=$6, rounding=$7, total=$8,
			notes=$9, your_reference=$10, our_reference=$11,
			revenue_account=$12, payment_account=$13, updated_at=NOW()
		WHERE invoice_id=$14 AND status='draft'
	`
	_, err := r.db.Exec(query,
		invoice.CustomerID, invoice.InvoiceDate.Time, invoice.DueDate.Time, invoice.PaymentTermsDays,
		invoice.Subtotal, invoice.VATTotal, invoice.Rounding, invoice.Total,
		invoice.Notes, invoice.YourReference, invoice.OurReference,
		invoice.RevenueAccount, invoice.PaymentAccount, invoice.InvoiceID,
	)
	return err
}

func (r *invoiceRepository) UpdateInvoiceStatus(invoiceID int, status string) error {
	_, err := r.db.Exec("UPDATE invoices SET status=$1, updated_at=NOW() WHERE invoice_id=$2", status, invoiceID)
	return err
}

func (r *invoiceRepository) LinkSalesVoucher(invoiceID int, voucherID int) error {
	_, err := r.db.Exec("UPDATE invoices SET sales_voucher_id=$1, updated_at=NOW() WHERE invoice_id=$2", voucherID, invoiceID)
	return err
}

func (r *invoiceRepository) LinkPaymentVoucher(invoiceID int, voucherID int) error {
	_, err := r.db.Exec("UPDATE invoices SET payment_voucher_id=$1, updated_at=NOW() WHERE invoice_id=$2", voucherID, invoiceID)
	return err
}

func (r *invoiceRepository) MarkAsPaid(invoiceID int, amountPaid decimal.Decimal, paidAt string) error {
	_, err := r.db.Exec(
		"UPDATE invoices SET status='paid', amount_paid=$1, paid_at=$2, updated_at=NOW() WHERE invoice_id=$3",
		amountPaid, paidAt, invoiceID,
	)
	return err
}

func (r *invoiceRepository) MarkAsSent(invoiceID int, sentAt string) error {
	_, err := r.db.Exec(
		"UPDATE invoices SET status='sent', sent_at=$1, updated_at=NOW() WHERE invoice_id=$2",
		sentAt, invoiceID,
	)
	return err
}

func (r *invoiceRepository) MarkAsCancelled(invoiceID int, cancelledAt string) error {
	_, err := r.db.Exec(
		"UPDATE invoices SET status='cancelled', cancelled_at=$1, updated_at=NOW() WHERE invoice_id=$2",
		cancelledAt, invoiceID,
	)
	return err
}

func (r *invoiceRepository) DeleteInvoice(invoiceID int) error {
	_, err := r.db.Exec("DELETE FROM invoices WHERE invoice_id=$1 AND status='draft'", invoiceID)
	return err
}
