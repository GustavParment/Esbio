package service

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type AccountService struct {
	repository repository.AccountRepository
	validate   *validator.Validate
}

func NewAccountService(repo repository.AccountRepository) *AccountService {
	return &AccountService{
		repository: repo,
		validate:   validator.New(),
	}
}

// CreateAccount creates a new account
func (s *AccountService) CreateAccount(account *domain.Account) error {
	// Validate input
	if err := s.validate.Struct(account); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if account number already exists
	existingAccount, err := s.repository.GetAccountByNo(account.AccountNo)
	if err == nil && existingAccount != nil {
		return errors.New("account with this account number already exists")
	}

	// Validate account group (1-8 for BAS-planen)
	if account.AccountGroup < 1 || account.AccountGroup > 8 {
		return errors.New("account group must be between 1 and 8")
	}

	// Validate type
	if account.Type != "P&L" && account.Type != "BS" {
		return errors.New("type must be either 'P&L' (Resultat) or 'BS' (Balans)")
	}

	// Validate standard side
	if account.StandardSide != "Debit" && account.StandardSide != "Credit" {
		return errors.New("standard side must be either 'Debit' or 'Credit'")
	}

	// Create account
	err = s.repository.CreateAccount(account)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetAccountByNo retrieves an account by account number
func (s *AccountService) GetAccountByNo(accountNo int) (*domain.Account, error) {
	if accountNo <= 0 {
		return nil, errors.New("invalid account number")
	}

	account, err := s.repository.GetAccountByNo(accountNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

// GetAllAccounts retrieves all accounts
func (s *AccountService) GetAllAccounts() ([]*domain.Account, error) {
	accounts, err := s.repository.GetAllAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	return accounts, nil
}

// GetAccountsByGroup retrieves accounts by account group
func (s *AccountService) GetAccountsByGroup(accountGroup int) ([]*domain.Account, error) {
	if accountGroup < 1 || accountGroup > 8 {
		return nil, errors.New("account group must be between 1 and 8")
	}

	accounts, err := s.repository.GetAccountsByGroup(accountGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by group: %w", err)
	}

	return accounts, nil
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(account *domain.Account) error {
	// Validate input
	if err := s.validate.Struct(account); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if account exists
	existingAccount, err := s.repository.GetAccountByNo(account.AccountNo)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}
	if existingAccount == nil {
		return errors.New("account not found")
	}

	// Validate account group
	if account.AccountGroup < 1 || account.AccountGroup > 8 {
		return errors.New("account group must be between 1 and 8")
	}

	// Validate type
	if account.Type != "P&L" && account.Type != "BS" {
		return errors.New("type must be either 'P&L' (Resultat) or 'BS' (Balans)")
	}

	// Validate standard side
	if account.StandardSide != "Debit" && account.StandardSide != "Credit" {
		return errors.New("standard side must be either 'Debit' or 'Credit'")
	}

	// Update account
	err = s.repository.UpdateAccount(account)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// DeleteAccount deletes an account by account number
func (s *AccountService) DeleteAccount(accountNo int) error {
	if accountNo <= 0 {
		return errors.New("invalid account number")
	}

	// Check if account exists
	existingAccount, err := s.repository.GetAccountByNo(accountNo)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}
	if existingAccount == nil {
		return errors.New("account not found")
	}

	// Delete account
	err = s.repository.DeleteAccount(accountNo)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// GetLedger retrieves ledger entries for an account
func (s *AccountService) GetLedger(accountNo int, period string) ([]*domain.LedgerEntry, error) {
	if accountNo <= 0 {
		return nil, errors.New("invalid account number")
	}

	// Check if account exists
	existingAccount, err := s.repository.GetAccountByNo(accountNo)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	if existingAccount == nil {
		return nil, errors.New("account not found")
	}

	// Get ledger entries
	entries, err := s.repository.GetLedger(accountNo, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger: %w", err)
	}

	return entries, nil
}
