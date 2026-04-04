package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
)

type LineItemRepository interface {
	CreateLineItem(lineItem *domain.LineItem) error
	GetLineItemByID(lineID int) (*domain.LineItem, error)
	GetLineItemsByVoucherID(voucherID int) ([]*domain.LineItem, error)
	GetLineItemsByAccountNo(accountNo int) ([]*domain.LineItem, error)
	UpdateLineItem(lineItem *domain.LineItem) error
	DeleteLineItem(lineID int) error
	DeleteLineItemsByVoucherID(voucherID int) error
}

type lineItemRepository struct {
	db *sql.DB
}

func NewLineItemRepository(db *sql.DB) LineItemRepository {
	return &lineItemRepository{db: db}
}

func (r *lineItemRepository) CreateLineItem(lineItem *domain.LineItem) error {
	query := `
		INSERT INTO line_items (voucher_id, account_no, debit_amount, credit_amount, tax_code, project_id, cost_center_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING line_id
	`
	err := r.db.QueryRow(query,
		lineItem.VoucherID,
		lineItem.AccountNo,
		lineItem.DebitAmount,
		lineItem.CreditAmount,
		lineItem.TaxCode,
		lineItem.ProjectID,
		lineItem.CostCenterID,
	).Scan(&lineItem.LineID)
	if err != nil {
		return fmt.Errorf("failed to create line item: %w", err)
	}

	return nil
}

func (r *lineItemRepository) GetLineItemByID(lineID int) (*domain.LineItem, error) {
	query := `
		SELECT line_id, voucher_id, account_no, debit_amount, credit_amount, COALESCE(tax_code, 0), COALESCE(project_id, 0), COALESCE(cost_center_id, 0)
		FROM line_items
		WHERE line_id = $1
	`
	lineItem := &domain.LineItem{}
	err := r.db.QueryRow(query, lineID).Scan(
		&lineItem.LineID,
		&lineItem.VoucherID,
		&lineItem.AccountNo,
		&lineItem.DebitAmount,
		&lineItem.CreditAmount,
		&lineItem.TaxCode,
		&lineItem.ProjectID,
		&lineItem.CostCenterID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("line item not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get line item: %w", err)
	}

	return lineItem, nil
}

func (r *lineItemRepository) GetLineItemsByVoucherID(voucherID int) ([]*domain.LineItem, error) {
	query := `
		SELECT line_id, voucher_id, account_no, debit_amount, credit_amount, COALESCE(tax_code, 0), COALESCE(project_id, 0), COALESCE(cost_center_id, 0)
		FROM line_items
		WHERE voucher_id = $1
		ORDER BY line_id
	`
	rows, err := r.db.Query(query, voucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items by voucher: %w", err)
	}
	defer rows.Close()

	var lineItems []*domain.LineItem
	for rows.Next() {
		lineItem := &domain.LineItem{}
		err := rows.Scan(
			&lineItem.LineID,
			&lineItem.VoucherID,
			&lineItem.AccountNo,
			&lineItem.DebitAmount,
			&lineItem.CreditAmount,
			&lineItem.TaxCode,
			&lineItem.ProjectID,
			&lineItem.CostCenterID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line item: %w", err)
		}
		lineItems = append(lineItems, lineItem)
	}

	return lineItems, nil
}

func (r *lineItemRepository) GetLineItemsByAccountNo(accountNo int) ([]*domain.LineItem, error) {
	query := `
		SELECT line_id, voucher_id, account_no, debit_amount, credit_amount, COALESCE(tax_code, 0), COALESCE(project_id, 0), COALESCE(cost_center_id, 0)
		FROM line_items
		WHERE account_no = $1
		ORDER BY line_id
	`
	rows, err := r.db.Query(query, accountNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items by account: %w", err)
	}
	defer rows.Close()

	var lineItems []*domain.LineItem
	for rows.Next() {
		lineItem := &domain.LineItem{}
		err := rows.Scan(
			&lineItem.LineID,
			&lineItem.VoucherID,
			&lineItem.AccountNo,
			&lineItem.DebitAmount,
			&lineItem.CreditAmount,
			&lineItem.TaxCode,
			&lineItem.ProjectID,
			&lineItem.CostCenterID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line item: %w", err)
		}
		lineItems = append(lineItems, lineItem)
	}

	return lineItems, nil
}

func (r *lineItemRepository) UpdateLineItem(lineItem *domain.LineItem) error {
	query := `
		UPDATE line_items
		SET voucher_id = $1, account_no = $2, debit_amount = $3, credit_amount = $4, tax_code = $5, project_id = $6, cost_center_id = $7
		WHERE line_id = $8
	`
	_, err := r.db.Exec(query,
		lineItem.VoucherID,
		lineItem.AccountNo,
		lineItem.DebitAmount,
		lineItem.CreditAmount,
		lineItem.TaxCode,
		lineItem.ProjectID,
		lineItem.CostCenterID,
		lineItem.LineID,
	)
	if err != nil {
		return fmt.Errorf("failed to update line item: %w", err)
	}

	return nil
}

func (r *lineItemRepository) DeleteLineItem(lineID int) error {
	query := `DELETE FROM line_items WHERE line_id = $1`
	_, err := r.db.Exec(query, lineID)
	if err != nil {
		return fmt.Errorf("failed to delete line item: %w", err)
	}

	return nil
}

func (r *lineItemRepository) DeleteLineItemsByVoucherID(voucherID int) error {
	query := `DELETE FROM line_items WHERE voucher_id = $1`
	_, err := r.db.Exec(query, voucherID)
	if err != nil {
		return fmt.Errorf("failed to delete line items by voucher: %w", err)
	}

	return nil
}
