package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
	"strings"
)

type BankTransactionRepository interface {
	BulkUpsert(txns []*domain.BankTransaction) (int, error)
	GetByCompanyID(companyID int, status string, limit, offset int) ([]*domain.BankTransaction, error)
	GetPendingForReview(companyID int) ([]*domain.BankTransaction, error)
	GetByID(txnID int) (*domain.BankTransaction, error)
	UpdateImportStatus(txnID int, status string, voucherID *int) error
	UpdateSuggestion(txnID int, accountNo int, description string) error
	CountByCompanyAndStatus(companyID int, status string) (int, error)
	FindPotentialDuplicates(companyID int, date string, amount float64) ([]*domain.BankTransaction, error)
}

type bankTransactionRepository struct {
	db *sql.DB
}

func NewBankTransactionRepository(db *sql.DB) BankTransactionRepository {
	return &bankTransactionRepository{db: db}
}

func (r *bankTransactionRepository) BulkUpsert(txns []*domain.BankTransaction) (int, error) {
	if len(txns) == 0 {
		return 0, nil
	}

	inserted := 0
	for _, txn := range txns {
		query := `
			INSERT INTO bank_transactions (bank_account_id, company_id, tink_transaction_id, booked_date, amount, currency, description_original, description_display, merchant_name, merchant_category_code, tink_category_id, status, import_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'pending')
			ON CONFLICT (tink_transaction_id) DO UPDATE SET
				status = EXCLUDED.status,
				description_display = EXCLUDED.description_display,
				merchant_name = EXCLUDED.merchant_name,
				updated_at = CURRENT_TIMESTAMP
			RETURNING (xmax = 0) AS is_insert
		`
		var isInsert bool
		err := r.db.QueryRow(query,
			txn.BankAccountID, txn.CompanyID, txn.TinkTransactionID,
			txn.BookedDate, txn.Amount, txn.Currency,
			txn.DescriptionOriginal, txn.DescriptionDisplay,
			txn.MerchantName, txn.MerchantCategoryCode, txn.TinkCategoryID,
			txn.Status,
		).Scan(&isInsert)
		if err != nil {
			return inserted, fmt.Errorf("failed to upsert transaction %s: %w", txn.TinkTransactionID, err)
		}
		if isInsert {
			inserted++
		}
	}
	return inserted, nil
}

func (r *bankTransactionRepository) GetByCompanyID(companyID int, status string, limit, offset int) ([]*domain.BankTransaction, error) {
	conditions := []string{"company_id = $1"}
	args := []interface{}{companyID}
	argIdx := 2

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("import_status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT bank_transaction_id, bank_account_id, company_id, tink_transaction_id, booked_date, amount, currency, description_original, description_display, merchant_name, merchant_category_code, tink_category_id, status, import_status, suggested_account_no, suggested_description, voucher_id, created_at, updated_at
		FROM bank_transactions
		WHERE %s
		ORDER BY booked_date DESC, bank_transaction_id DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), argIdx, argIdx+1)
	args = append(args, limit, offset)

	return r.scanTransactions(r.db.Query(query, args...))
}

func (r *bankTransactionRepository) GetPendingForReview(companyID int) ([]*domain.BankTransaction, error) {
	query := `
		SELECT bank_transaction_id, bank_account_id, company_id, tink_transaction_id, booked_date, amount, currency, description_original, description_display, merchant_name, merchant_category_code, tink_category_id, status, import_status, suggested_account_no, suggested_description, voucher_id, created_at, updated_at
		FROM bank_transactions
		WHERE company_id = $1 AND import_status IN ('pending', 'suggested')
		ORDER BY booked_date DESC, bank_transaction_id DESC
	`
	return r.scanTransactions(r.db.Query(query, companyID))
}

func (r *bankTransactionRepository) GetByID(txnID int) (*domain.BankTransaction, error) {
	query := `
		SELECT bank_transaction_id, bank_account_id, company_id, tink_transaction_id, booked_date, amount, currency, description_original, description_display, merchant_name, merchant_category_code, tink_category_id, status, import_status, suggested_account_no, suggested_description, voucher_id, created_at, updated_at
		FROM bank_transactions WHERE bank_transaction_id = $1
	`
	txn := &domain.BankTransaction{}
	err := r.db.QueryRow(query, txnID).Scan(
		&txn.BankTransactionID, &txn.BankAccountID, &txn.CompanyID,
		&txn.TinkTransactionID, &txn.BookedDate, &txn.Amount, &txn.Currency,
		&txn.DescriptionOriginal, &txn.DescriptionDisplay,
		&txn.MerchantName, &txn.MerchantCategoryCode, &txn.TinkCategoryID,
		&txn.Status, &txn.ImportStatus,
		&txn.SuggestedAccountNo, &txn.SuggestedDescription,
		&txn.VoucherID, &txn.CreatedAt, &txn.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bank transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bank transaction: %w", err)
	}
	return txn, nil
}

func (r *bankTransactionRepository) UpdateImportStatus(txnID int, status string, voucherID *int) error {
	query := `UPDATE bank_transactions SET import_status = $1, voucher_id = $2, updated_at = CURRENT_TIMESTAMP WHERE bank_transaction_id = $3`
	_, err := r.db.Exec(query, status, voucherID, txnID)
	return err
}

func (r *bankTransactionRepository) UpdateSuggestion(txnID int, accountNo int, description string) error {
	query := `UPDATE bank_transactions SET suggested_account_no = $1, suggested_description = $2, import_status = 'suggested', updated_at = CURRENT_TIMESTAMP WHERE bank_transaction_id = $3`
	_, err := r.db.Exec(query, accountNo, description, txnID)
	return err
}

func (r *bankTransactionRepository) CountByCompanyAndStatus(companyID int, status string) (int, error) {
	query := `SELECT COUNT(*) FROM bank_transactions WHERE company_id = $1 AND import_status = $2`
	var count int
	err := r.db.QueryRow(query, companyID, status).Scan(&count)
	return count, err
}

func (r *bankTransactionRepository) FindPotentialDuplicates(companyID int, date string, amount float64) ([]*domain.BankTransaction, error) {
	query := `
		SELECT bank_transaction_id, bank_account_id, company_id, tink_transaction_id, booked_date, amount, currency, description_original, description_display, merchant_name, merchant_category_code, tink_category_id, status, import_status, suggested_account_no, suggested_description, voucher_id, created_at, updated_at
		FROM bank_transactions
		WHERE company_id = $1 AND booked_date = $2 AND ABS(amount - $3) < 0.01
	`
	return r.scanTransactions(r.db.Query(query, companyID, date, amount))
}

func (r *bankTransactionRepository) scanTransactions(rows *sql.Rows, err error) ([]*domain.BankTransaction, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to query bank transactions: %w", err)
	}
	defer rows.Close()

	var txns []*domain.BankTransaction
	for rows.Next() {
		txn := &domain.BankTransaction{}
		if err := rows.Scan(
			&txn.BankTransactionID, &txn.BankAccountID, &txn.CompanyID,
			&txn.TinkTransactionID, &txn.BookedDate, &txn.Amount, &txn.Currency,
			&txn.DescriptionOriginal, &txn.DescriptionDisplay,
			&txn.MerchantName, &txn.MerchantCategoryCode, &txn.TinkCategoryID,
			&txn.Status, &txn.ImportStatus,
			&txn.SuggestedAccountNo, &txn.SuggestedDescription,
			&txn.VoucherID, &txn.CreatedAt, &txn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bank transaction: %w", err)
		}
		txns = append(txns, txn)
	}
	return txns, nil
}
