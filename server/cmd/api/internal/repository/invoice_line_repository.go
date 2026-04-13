package repository

import (
	"database/sql"
	"esbio/api/internal/domain"
	"fmt"
)

type InvoiceLineRepository interface {
	CreateLine(line *domain.InvoiceLine) error
	GetLinesByInvoiceID(invoiceID int) ([]*domain.InvoiceLine, error)
	UpdateLine(line *domain.InvoiceLine) error
	DeleteLine(lineID int) error
	DeleteLinesByInvoiceID(invoiceID int) error
}

type invoiceLineRepository struct {
	db *sql.DB
}

func NewInvoiceLineRepository(db *sql.DB) InvoiceLineRepository {
	return &invoiceLineRepository{db: db}
}

func (r *invoiceLineRepository) CreateLine(line *domain.InvoiceLine) error {
	query := `
		INSERT INTO invoice_lines (
			invoice_id, sort_order, description, quantity, unit,
			unit_price, discount_percent, vat_rate, line_total, vat_amount, account_no
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING line_id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		line.InvoiceID, line.SortOrder, line.Description, line.Quantity, line.Unit,
		line.UnitPrice, line.DiscountPercent, line.VATRate, line.LineTotal, line.VATAmount, line.AccountNo,
	).Scan(&line.LineID, &line.CreatedAt, &line.UpdatedAt)
}

func (r *invoiceLineRepository) GetLinesByInvoiceID(invoiceID int) ([]*domain.InvoiceLine, error) {
	query := `
		SELECT line_id, invoice_id, sort_order, description, quantity, COALESCE(unit,'st'),
			unit_price, discount_percent, vat_rate, line_total, vat_amount, account_no,
			created_at, updated_at
		FROM invoice_lines WHERE invoice_id = $1 ORDER BY sort_order, line_id
	`
	rows, err := r.db.Query(query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice lines: %w", err)
	}
	defer rows.Close()

	var lines []*domain.InvoiceLine
	for rows.Next() {
		l := &domain.InvoiceLine{}
		if err := rows.Scan(
			&l.LineID, &l.InvoiceID, &l.SortOrder, &l.Description, &l.Quantity, &l.Unit,
			&l.UnitPrice, &l.DiscountPercent, &l.VATRate, &l.LineTotal, &l.VATAmount, &l.AccountNo,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan invoice line: %w", err)
		}
		lines = append(lines, l)
	}
	return lines, nil
}

func (r *invoiceLineRepository) UpdateLine(line *domain.InvoiceLine) error {
	query := `
		UPDATE invoice_lines SET
			sort_order=$1, description=$2, quantity=$3, unit=$4,
			unit_price=$5, discount_percent=$6, vat_rate=$7,
			line_total=$8, vat_amount=$9, account_no=$10, updated_at=NOW()
		WHERE line_id=$11
	`
	_, err := r.db.Exec(query,
		line.SortOrder, line.Description, line.Quantity, line.Unit,
		line.UnitPrice, line.DiscountPercent, line.VATRate,
		line.LineTotal, line.VATAmount, line.AccountNo, line.LineID,
	)
	return err
}

func (r *invoiceLineRepository) DeleteLine(lineID int) error {
	_, err := r.db.Exec("DELETE FROM invoice_lines WHERE line_id=$1", lineID)
	return err
}

func (r *invoiceLineRepository) DeleteLinesByInvoiceID(invoiceID int) error {
	_, err := r.db.Exec("DELETE FROM invoice_lines WHERE invoice_id=$1", invoiceID)
	return err
}
