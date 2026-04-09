package repository

import (
	"esbio/api/internal/domain"
	"database/sql"
	"fmt"
)

type AccountRepository interface {
	CreateAccount(account *domain.Account) error
	GetAccountByNo(accountNo int) (*domain.Account, error)
	GetAllAccounts() ([]*domain.Account, error)
	GetAccountsByGroup(accountGroup int) ([]*domain.Account, error)
	UpdateAccount(account *domain.Account) error
	DeleteAccount(accountNo int) error
	GetLedger(accountNo int, period string) ([]*domain.LedgerEntry, error)
}

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) CreateAccount(account *domain.Account) error {
	query := `
		INSERT INTO accounts (account_no, account_name, account_group, tax_standard, type, standard_side)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query,
		account.AccountNo,
		account.AccountName,
		account.AccountGroup,
		account.TaxStandard,
		account.Type,
		account.StandardSide,
	)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (r *accountRepository) GetAccountByNo(accountNo int) (*domain.Account, error) {
	query := `
		SELECT account_no, account_name, account_group, tax_standard, type, standard_side
		FROM accounts
		WHERE account_no = $1
	`
	account := &domain.Account{}
	err := r.db.QueryRow(query, accountNo).Scan(
		&account.AccountNo,
		&account.AccountName,
		&account.AccountGroup,
		&account.TaxStandard,
		&account.Type,
		&account.StandardSide,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

func (r *accountRepository) GetAllAccounts() ([]*domain.Account, error) {
	query := `
		SELECT account_no, account_name, account_group, tax_standard, type, standard_side
		FROM accounts
		ORDER BY account_no
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]*domain.Account, 0)
	for rows.Next() {
		account := &domain.Account{}
		err := rows.Scan(
			&account.AccountNo,
			&account.AccountName,
			&account.AccountGroup,
			&account.TaxStandard,
			&account.Type,
			&account.StandardSide,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *accountRepository) GetAccountsByGroup(accountGroup int) ([]*domain.Account, error) {
	query := `
		SELECT account_no, account_name, account_group, tax_standard, type, standard_side
		FROM accounts
		WHERE account_group = $1
		ORDER BY account_no
	`
	rows, err := r.db.Query(query, accountGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by group: %w", err)
	}
	defer rows.Close()

	accounts := make([]*domain.Account, 0)
	for rows.Next() {
		account := &domain.Account{}
		err := rows.Scan(
			&account.AccountNo,
			&account.AccountName,
			&account.AccountGroup,
			&account.TaxStandard,
			&account.Type,
			&account.StandardSide,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *accountRepository) UpdateAccount(account *domain.Account) error {
	query := `
		UPDATE accounts
		SET account_name = $1, account_group = $2, tax_standard = $3, type = $4, standard_side = $5
		WHERE account_no = $6
	`
	_, err := r.db.Exec(query,
		account.AccountName,
		account.AccountGroup,
		account.TaxStandard,
		account.Type,
		account.StandardSide,
		account.AccountNo,
	)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

func (r *accountRepository) DeleteAccount(accountNo int) error {
	query := `DELETE FROM accounts WHERE account_no = $1`
	_, err := r.db.Exec(query, accountNo)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

func (r *accountRepository) GetLedger(accountNo int, period string) ([]*domain.LedgerEntry, error) {
	// Calculate opening balance from prior periods when filtering
	var runningBalance float64 = 0
	if period != "" {
		obQuery := `
			SELECT COALESCE(SUM(l.debit_amount - l.credit_amount), 0)
			FROM line_items l
			INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
			WHERE l.account_no = $1
			  AND v.period < $2
			  AND v.corrected_by_voucher_id IS NULL
		`
		if err := r.db.QueryRow(obQuery, accountNo, period).Scan(&runningBalance); err != nil {
			return nil, fmt.Errorf("failed to get opening balance: %w", err)
		}
	}

	query := `
		SELECT
			v.date,
			v.voucher_id,
			v.voucher_number,
			v.description,
			v.reference,
			l.debit_amount,
			l.credit_amount
		FROM line_items l
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		WHERE l.account_no = $1
			AND ($2 = '' OR v.period = $2)
			AND v.corrected_by_voucher_id IS NULL
		ORDER BY v.date ASC, v.voucher_number ASC
	`

	rows, err := r.db.Query(query, accountNo, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*domain.LedgerEntry, 0)

	for rows.Next() {
		entry := &domain.LedgerEntry{}
		err := rows.Scan(
			&entry.Date.Time,
			&entry.VoucherID,
			&entry.VoucherNumber,
			&entry.Description,
			&entry.Reference,
			&entry.DebitAmount,
			&entry.CreditAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}

		// Calculate running balance (Debit increases, Credit decreases)
		runningBalance += entry.DebitAmount - entry.CreditAmount
		entry.Balance = runningBalance

		entries = append(entries, entry)
	}

	return entries, nil
}