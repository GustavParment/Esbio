package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
)

type BankAccountRepository interface {
	Create(account *domain.BankAccount) error
	GetByID(bankAccountID int) (*domain.BankAccount, error)
	GetByCompanyID(companyID int) ([]*domain.BankAccount, error)
	GetByConnectionID(connectionID int) ([]*domain.BankAccount, error)
	GetByTinkAccountID(tinkAccountID string) (*domain.BankAccount, error)
	UpdateMapping(bankAccountID int, mappedAccountNo int) error
	UpdateBalance(bankAccountID int, balance float64) error
	Delete(bankAccountID int) error
}

type bankAccountRepository struct {
	db *sql.DB
}

func NewBankAccountRepository(db *sql.DB) BankAccountRepository {
	return &bankAccountRepository{db: db}
}

func (r *bankAccountRepository) Create(account *domain.BankAccount) error {
	query := `
		INSERT INTO bank_accounts (connection_id, company_id, tink_account_id, account_name, iban, account_number, currency, balance_amount, balance_updated_at, mapped_account_no)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tink_account_id) DO UPDATE SET
			account_name = EXCLUDED.account_name,
			balance_amount = EXCLUDED.balance_amount,
			balance_updated_at = EXCLUDED.balance_updated_at,
			updated_at = CURRENT_TIMESTAMP
		RETURNING bank_account_id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		account.ConnectionID, account.CompanyID, account.TinkAccountID,
		account.AccountName, account.IBAN, account.AccountNumber,
		account.Currency, account.BalanceAmount, account.BalanceUpdatedAt,
		account.MappedAccountNo,
	).Scan(&account.BankAccountID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *bankAccountRepository) GetByID(bankAccountID int) (*domain.BankAccount, error) {
	query := `
		SELECT bank_account_id, connection_id, company_id, tink_account_id, account_name, iban, account_number, currency, balance_amount, balance_updated_at, mapped_account_no, created_at, updated_at
		FROM bank_accounts WHERE bank_account_id = $1
	`
	a := &domain.BankAccount{}
	err := r.db.QueryRow(query, bankAccountID).Scan(
		&a.BankAccountID, &a.ConnectionID, &a.CompanyID, &a.TinkAccountID,
		&a.AccountName, &a.IBAN, &a.AccountNumber, &a.Currency,
		&a.BalanceAmount, &a.BalanceUpdatedAt, &a.MappedAccountNo,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bank account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bank account: %w", err)
	}
	return a, nil
}

func (r *bankAccountRepository) GetByCompanyID(companyID int) ([]*domain.BankAccount, error) {
	query := `
		SELECT bank_account_id, connection_id, company_id, tink_account_id, account_name, iban, account_number, currency, balance_amount, balance_updated_at, mapped_account_no, created_at, updated_at
		FROM bank_accounts WHERE company_id = $1
		ORDER BY account_name
	`
	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.BankAccount
	for rows.Next() {
		a := &domain.BankAccount{}
		if err := rows.Scan(
			&a.BankAccountID, &a.ConnectionID, &a.CompanyID, &a.TinkAccountID,
			&a.AccountName, &a.IBAN, &a.AccountNumber, &a.Currency,
			&a.BalanceAmount, &a.BalanceUpdatedAt, &a.MappedAccountNo,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bank account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *bankAccountRepository) GetByConnectionID(connectionID int) ([]*domain.BankAccount, error) {
	query := `
		SELECT bank_account_id, connection_id, company_id, tink_account_id, account_name, iban, account_number, currency, balance_amount, balance_updated_at, mapped_account_no, created_at, updated_at
		FROM bank_accounts WHERE connection_id = $1
		ORDER BY account_name
	`
	rows, err := r.db.Query(query, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts by connection: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.BankAccount
	for rows.Next() {
		a := &domain.BankAccount{}
		if err := rows.Scan(
			&a.BankAccountID, &a.ConnectionID, &a.CompanyID, &a.TinkAccountID,
			&a.AccountName, &a.IBAN, &a.AccountNumber, &a.Currency,
			&a.BalanceAmount, &a.BalanceUpdatedAt, &a.MappedAccountNo,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bank account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *bankAccountRepository) GetByTinkAccountID(tinkAccountID string) (*domain.BankAccount, error) {
	query := `
		SELECT bank_account_id, connection_id, company_id, tink_account_id, account_name, iban, account_number, currency, balance_amount, balance_updated_at, mapped_account_no, created_at, updated_at
		FROM bank_accounts WHERE tink_account_id = $1
	`
	a := &domain.BankAccount{}
	err := r.db.QueryRow(query, tinkAccountID).Scan(
		&a.BankAccountID, &a.ConnectionID, &a.CompanyID, &a.TinkAccountID,
		&a.AccountName, &a.IBAN, &a.AccountNumber, &a.Currency,
		&a.BalanceAmount, &a.BalanceUpdatedAt, &a.MappedAccountNo,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bank account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bank account by tink ID: %w", err)
	}
	return a, nil
}

func (r *bankAccountRepository) UpdateMapping(bankAccountID int, mappedAccountNo int) error {
	query := `UPDATE bank_accounts SET mapped_account_no = $1, updated_at = CURRENT_TIMESTAMP WHERE bank_account_id = $2`
	_, err := r.db.Exec(query, mappedAccountNo, bankAccountID)
	return err
}

func (r *bankAccountRepository) UpdateBalance(bankAccountID int, balance float64) error {
	query := `UPDATE bank_accounts SET balance_amount = $1, balance_updated_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE bank_account_id = $2`
	_, err := r.db.Exec(query, balance, bankAccountID)
	return err
}

func (r *bankAccountRepository) Delete(bankAccountID int) error {
	query := `DELETE FROM bank_accounts WHERE bank_account_id = $1`
	_, err := r.db.Exec(query, bankAccountID)
	return err
}
