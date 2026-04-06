package service

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type VoucherService struct {
	repository         repository.VoucherRepository
	lineItemRepository repository.LineItemRepository
	validate           *validator.Validate
}

func NewVoucherService(repo repository.VoucherRepository, lineItemRepo repository.LineItemRepository) *VoucherService {
	return &VoucherService{
		repository:         repo,
		lineItemRepository: lineItemRepo,
		validate:           validator.New(),
	}
}

// CreateVoucher creates a new voucher
func (s *VoucherService) CreateVoucher(voucher *domain.Voucher) error {
	// Validate input
	if err := s.validate.Struct(voucher); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate description
	if voucher.Description == "" {
		return errors.New("description is required")
	}

	// Validate period format (e.g., "2025-01")
	if len(voucher.Period) != 7 {
		return errors.New("period must be in format YYYY-MM (e.g., '2025-01')")
	}

	// Validate created by
	if voucher.CreatedBy <= 0 {
		return errors.New("invalid user ID")
	}

	// Create voucher
	err := s.repository.CreateVoucher(voucher)
	if err != nil {
		return fmt.Errorf("failed to create voucher: %w", err)
	}

	return nil
}

// GetVoucherByID retrieves a voucher by ID, including its line items
func (s *VoucherService) GetVoucherByID(voucherID int) (*domain.Voucher, error) {
	if voucherID <= 0 {
		return nil, errors.New("invalid voucher ID")
	}

	voucher, err := s.repository.GetVoucherByID(voucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get voucher: %w", err)
	}

	// Get line items for this voucher
	lineItems, err := s.lineItemRepository.GetLineItemsByVoucherID(voucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items: %w", err)
	}

	// Convert []*LineItem to []LineItem
	voucher.Lines = make([]domain.LineItem, len(lineItems))
	for i, item := range lineItems {
		voucher.Lines[i] = *item
	}

	return voucher, nil
}

func (s *VoucherService) GetVoucherByNumber(companyID int, voucherNumber int) (*domain.Voucher, error) {
	voucher, err := s.repository.GetVoucherByNumber(companyID, voucherNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get voucher: %w", err)
	}

	lineItems, err := s.lineItemRepository.GetLineItemsByVoucherID(voucher.VoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items: %w", err)
	}

	voucher.Lines = make([]domain.LineItem, len(lineItems))
	for i, item := range lineItems {
		voucher.Lines[i] = *item
	}

	return voucher, nil
}

// GetAllVouchers retrieves all vouchers
func (s *VoucherService) GetAllVouchers() ([]*domain.Voucher, error) {
	vouchers, err := s.repository.GetAllVouchers()
	if err != nil {
		return nil, fmt.Errorf("failed to get vouchers: %w", err)
	}

	return vouchers, nil
}

// GetVouchersByPeriod retrieves vouchers by period for a specific company
func (s *VoucherService) GetVouchersByPeriod(companyID int, period string) ([]*domain.Voucher, error) {
	if len(period) != 7 {
		return nil, errors.New("period must be in format YYYY-MM (e.g., '2025-01')")
	}

	vouchers, err := s.repository.GetVouchersByPeriod(companyID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get vouchers by period: %w", err)
	}

	return vouchers, nil
}

// GetVouchersByCompanyID retrieves vouchers by company
func (s *VoucherService) GetVouchersByCompanyID(companyID int) ([]*domain.Voucher, error) {
	if companyID <= 0 {
		return nil, errors.New("invalid company ID")
	}

	vouchers, err := s.repository.GetVouchersByCompanyID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vouchers by company: %w", err)
	}

	return vouchers, nil
}

// GetAllPeriods retrieves all unique periods from vouchers for a specific company
func (s *VoucherService) GetAllPeriods(companyID int) ([]string, error) {
	periods, err := s.repository.GetAllPeriods(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get periods: %w", err)
	}

	return periods, nil
}

// UpdateVoucher updates an existing voucher
func (s *VoucherService) UpdateVoucher(voucher *domain.Voucher) error {
	// Validate input
	if err := s.validate.Struct(voucher); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if voucher exists
	existingVoucher, err := s.repository.GetVoucherByID(voucher.VoucherID)
	if err != nil {
		return fmt.Errorf("voucher not found: %w", err)
	}
	if existingVoucher == nil {
		return errors.New("voucher not found")
	}

	// Validate description
	if voucher.Description == "" {
		return errors.New("description is required")
	}

	// Validate period format
	if len(voucher.Period) != 7 {
		return errors.New("period must be in format YYYY-MM (e.g., '2025-01')")
	}

	// Validate created by
	if voucher.CreatedBy <= 0 {
		return errors.New("invalid user ID")
	}

	// Update voucher
	err = s.repository.UpdateVoucher(voucher)
	if err != nil {
		return fmt.Errorf("failed to update voucher: %w", err)
	}

	return nil
}

// DeleteVoucher deletes a voucher by ID (also deletes associated line items)
func (s *VoucherService) DeleteVoucher(voucherID int) error {
	if voucherID <= 0 {
		return errors.New("invalid voucher ID")
	}

	// Check if voucher exists
	existingVoucher, err := s.repository.GetVoucherByID(voucherID)
	if err != nil {
		return fmt.Errorf("voucher not found: %w", err)
	}
	if existingVoucher == nil {
		return errors.New("voucher not found")
	}

	// Delete associated line items first
	err = s.lineItemRepository.DeleteLineItemsByVoucherID(voucherID)
	if err != nil {
		return fmt.Errorf("failed to delete line items: %w", err)
	}

	// Delete voucher
	err = s.repository.DeleteVoucher(voucherID)
	if err != nil {
		return fmt.Errorf("failed to delete voucher: %w", err)
	}

	return nil
}

// ValidateVoucherBalance validates that debit and credit balance for a voucher
func (s *VoucherService) ValidateVoucherBalance(voucherID int) (bool, error) {
	lineItems, err := s.lineItemRepository.GetLineItemsByVoucherID(voucherID)
	if err != nil {
		return false, fmt.Errorf("failed to get line items: %w", err)
	}

	var totalDebit, totalCredit float64
	for _, item := range lineItems {
		totalDebit += item.DebitAmount
		totalCredit += item.CreditAmount
	}

	// Check if debit equals credit (with small tolerance for floating point)
	balanced := (totalDebit - totalCredit) < 0.01 && (totalDebit - totalCredit) > -0.01

	return balanced, nil
}

// CreateCorrectionVoucher creates a correction voucher that reverses the original voucher
func (s *VoucherService) CreateCorrectionVoucher(originalVoucherID int, userID int) (*domain.Voucher, error) {
	if originalVoucherID <= 0 {
		return nil, errors.New("invalid voucher ID")
	}
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Get the original voucher
	originalVoucher, err := s.repository.GetVoucherByID(originalVoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original voucher: %w", err)
	}

	// Check if voucher is already corrected
	if originalVoucher.CorrectedByVoucherID != nil {
		return nil, errors.New("voucher has already been corrected")
	}

	// Get original line items
	originalLineItems, err := s.lineItemRepository.GetLineItemsByVoucherID(originalVoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original line items: %w", err)
	}

	// Create correction voucher with reversed amounts
	correctionVoucher := &domain.Voucher{
		Date:        domain.FlexibleDate{Time: originalVoucher.Date.Time},
		Description: fmt.Sprintf("Rättelse av verifikat #%d: %s", originalVoucher.VoucherNumber, originalVoucher.Description),
		Reference:   originalVoucher.Reference,
		TotalAmount: originalVoucher.TotalAmount,
		Period:      originalVoucher.Period,
		CreatedBy:   userID,
		CompanyID:   originalVoucher.CompanyID,
	}

	// Create the correction voucher in database
	err = s.repository.CreateCorrectionVoucher(correctionVoucher, originalVoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to create correction voucher: %w", err)
	}

	// Create reversed line items (swap debit and credit)
	for _, item := range originalLineItems {
		reversedItem := &domain.LineItem{
			VoucherID:    correctionVoucher.VoucherID,
			AccountNo:    item.AccountNo,
			DebitAmount:  item.CreditAmount,  // Swap: original credit becomes debit
			CreditAmount: item.DebitAmount,   // Swap: original debit becomes credit
			TaxCode:      item.TaxCode,
		}
		err = s.lineItemRepository.CreateLineItem(reversedItem)
		if err != nil {
			return nil, fmt.Errorf("failed to create correction line item: %w", err)
		}
	}

	// Mark the original voucher as corrected
	err = s.repository.MarkVoucherAsCorrected(originalVoucherID, correctionVoucher.VoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark original voucher as corrected: %w", err)
	}

	return correctionVoucher, nil
}

// CreateCorrectionWithChanges creates ONE new corrected voucher and marks the original as corrected
func (s *VoucherService) CreateCorrectionWithChanges(
	originalVoucherID int,
	userID int,
	newDate string,
	newDescription string,
	newReference string,
	newPeriod string,
	newLineItems []domain.LineItem,
) (*domain.Voucher, error) {
	if originalVoucherID <= 0 {
		return nil, errors.New("invalid voucher ID")
	}
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Get the original voucher
	originalVoucher, err := s.repository.GetVoucherByID(originalVoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original voucher: %w", err)
	}

	// Check if voucher is already corrected
	if originalVoucher.CorrectedByVoucherID != nil {
		return nil, errors.New("voucher has already been corrected")
	}

	// Calculate total from new line items
	var newTotal float64
	for _, item := range newLineItems {
		if item.DebitAmount > 0 {
			newTotal += item.DebitAmount
		}
	}

	// Parse new date
	parsedDate, err := time.Parse("2006-01-02", newDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create NEW CORRECTED voucher (with updated values)
	newVoucher := &domain.Voucher{
		Date:        domain.FlexibleDate{Time: parsedDate},
		Description: newDescription,
		Reference:   newReference,
		TotalAmount: newTotal,
		Period:      newPeriod,
		CreatedBy:   userID,
		CompanyID:   originalVoucher.CompanyID,
	}

	// Create the new voucher in database using CreateCorrectionVoucher to link it
	err = s.repository.CreateCorrectionVoucher(newVoucher, originalVoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new corrected voucher: %w", err)
	}

	// Create new line items (the corrected values)
	for _, item := range newLineItems {
		newItem := &domain.LineItem{
			VoucherID:    newVoucher.VoucherID,
			AccountNo:    item.AccountNo,
			DebitAmount:  item.DebitAmount,
			CreditAmount: item.CreditAmount,
			TaxCode:      item.TaxCode,
		}
		err = s.lineItemRepository.CreateLineItem(newItem)
		if err != nil {
			return nil, fmt.Errorf("failed to create new line item: %w", err)
		}
	}

	// Mark the original voucher as corrected (link to new voucher)
	err = s.repository.MarkVoucherAsCorrected(originalVoucherID, newVoucher.VoucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark original voucher as corrected: %w", err)
	}

	// Return the new corrected voucher (the user will be redirected here)
	return newVoucher, nil
}
