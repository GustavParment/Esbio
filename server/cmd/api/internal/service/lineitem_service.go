package service

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type LineItemService struct {
	repository repository.LineItemRepository
	validate   *validator.Validate
}

func NewLineItemService(repo repository.LineItemRepository) *LineItemService {
	return &LineItemService{
		repository: repo,
		validate:   validator.New(),
	}
}

// CreateLineItem creates a new line item
func (s *LineItemService) CreateLineItem(lineItem *domain.LineItem) error {
	// Validate input
	if err := s.validate.Struct(lineItem); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Business rule: Either debit or credit must be > 0, but not both
	if lineItem.DebitAmount > 0 && lineItem.CreditAmount > 0 {
		return errors.New("a line item cannot have both debit and credit amounts")
	}

	if lineItem.DebitAmount == 0 && lineItem.CreditAmount == 0 {
		return errors.New("a line item must have either debit or credit amount")
	}

	// Validate voucher ID
	if lineItem.VoucherID <= 0 {
		return errors.New("invalid voucher ID")
	}

	// Validate account number
	if lineItem.AccountNo <= 0 {
		return errors.New("invalid account number")
	}

	// Create line item
	err := s.repository.CreateLineItem(lineItem)
	if err != nil {
		return fmt.Errorf("failed to create line item: %w", err)
	}

	return nil
}

// GetLineItemByID retrieves a line item by ID
func (s *LineItemService) GetLineItemByID(lineID int) (*domain.LineItem, error) {
	if lineID <= 0 {
		return nil, errors.New("invalid line item ID")
	}

	lineItem, err := s.repository.GetLineItemByID(lineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line item: %w", err)
	}

	return lineItem, nil
}

// GetLineItemsByVoucherID retrieves all line items for a voucher
func (s *LineItemService) GetLineItemsByVoucherID(voucherID int) ([]*domain.LineItem, error) {
	if voucherID <= 0 {
		return nil, errors.New("invalid voucher ID")
	}

	lineItems, err := s.repository.GetLineItemsByVoucherID(voucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items by voucher: %w", err)
	}

	return lineItems, nil
}

// GetLineItemsByAccountNo retrieves all line items for an account
func (s *LineItemService) GetLineItemsByAccountNo(accountNo int) ([]*domain.LineItem, error) {
	if accountNo <= 0 {
		return nil, errors.New("invalid account number")
	}

	lineItems, err := s.repository.GetLineItemsByAccountNo(accountNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items by account: %w", err)
	}

	return lineItems, nil
}

// UpdateLineItem updates an existing line item
func (s *LineItemService) UpdateLineItem(lineItem *domain.LineItem) error {
	// Validate input
	if err := s.validate.Struct(lineItem); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if line item exists
	existingLineItem, err := s.repository.GetLineItemByID(lineItem.LineID)
	if err != nil {
		return fmt.Errorf("line item not found: %w", err)
	}
	if existingLineItem == nil {
		return errors.New("line item not found")
	}

	// Business rule: Either debit or credit must be > 0, but not both
	if lineItem.DebitAmount > 0 && lineItem.CreditAmount > 0 {
		return errors.New("a line item cannot have both debit and credit amounts")
	}

	if lineItem.DebitAmount == 0 && lineItem.CreditAmount == 0 {
		return errors.New("a line item must have either debit or credit amount")
	}

	// Validate voucher ID
	if lineItem.VoucherID <= 0 {
		return errors.New("invalid voucher ID")
	}

	// Validate account number
	if lineItem.AccountNo <= 0 {
		return errors.New("invalid account number")
	}

	// Update line item
	err = s.repository.UpdateLineItem(lineItem)
	if err != nil {
		return fmt.Errorf("failed to update line item: %w", err)
	}

	return nil
}

// DeleteLineItem deletes a line item by ID
func (s *LineItemService) DeleteLineItem(lineID int) error {
	if lineID <= 0 {
		return errors.New("invalid line item ID")
	}

	// Check if line item exists
	existingLineItem, err := s.repository.GetLineItemByID(lineID)
	if err != nil {
		return fmt.Errorf("line item not found: %w", err)
	}
	if existingLineItem == nil {
		return errors.New("line item not found")
	}

	// Delete line item
	err = s.repository.DeleteLineItem(lineID)
	if err != nil {
		return fmt.Errorf("failed to delete line item: %w", err)
	}

	return nil
}

// DeleteLineItemsByVoucherID deletes all line items for a voucher
func (s *LineItemService) DeleteLineItemsByVoucherID(voucherID int) error {
	if voucherID <= 0 {
		return errors.New("invalid voucher ID")
	}

	err := s.repository.DeleteLineItemsByVoucherID(voucherID)
	if err != nil {
		return fmt.Errorf("failed to delete line items by voucher: %w", err)
	}

	return nil
}
