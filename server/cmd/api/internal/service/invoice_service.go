package service

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// BAS account constants for auto-bokforing
const (
	AccountKundfordringar = 1510
	AccountUtgMoms25      = 2610
	AccountUtgMoms12      = 2611
	AccountUtgMoms6       = 2612
	AccountOresavrundning = 3740
)

type InvoiceService struct {
	invoiceRepo    repository.InvoiceRepository
	lineRepo       repository.InvoiceLineRepository
	settingsRepo   repository.InvoiceSettingsRepository
	customerRepo   repository.CustomerRepository
	voucherService *VoucherService
	lineItemRepo   repository.LineItemRepository
}

func NewInvoiceService(
	invoiceRepo repository.InvoiceRepository,
	lineRepo repository.InvoiceLineRepository,
	settingsRepo repository.InvoiceSettingsRepository,
	customerRepo repository.CustomerRepository,
	voucherService *VoucherService,
	lineItemRepo repository.LineItemRepository,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:    invoiceRepo,
		lineRepo:       lineRepo,
		settingsRepo:   settingsRepo,
		customerRepo:   customerRepo,
		voucherService: voucherService,
		lineItemRepo:   lineItemRepo,
	}
}

// CreateInvoiceRequest is the input for creating an invoice
type CreateInvoiceRequest struct {
	CustomerID       int                `json:"customer_id"`
	InvoiceDate      string             `json:"invoice_date"`
	PaymentTermsDays int                `json:"payment_terms_days"`
	Notes            string             `json:"notes"`
	YourReference    string             `json:"your_reference"`
	OurReference     string             `json:"our_reference"`
	RevenueAccount   int                `json:"revenue_account"`
	Lines            []CreateLineRequest `json:"lines"`
}

type CreateLineRequest struct {
	Description     string          `json:"description"`
	Quantity        decimal.Decimal `json:"quantity"`
	Unit            string          `json:"unit"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	VATRate         int             `json:"vat_rate"`
	AccountNo       int             `json:"account_no"`
}

// CreateInvoice creates a draft invoice with calculated totals
func (s *InvoiceService) CreateInvoice(companyID, userID int, req CreateInvoiceRequest) (*domain.Invoice, error) {
	if req.CustomerID <= 0 {
		return nil, errors.New("customer_id is required")
	}
	if len(req.Lines) == 0 {
		return nil, errors.New("at least one line is required")
	}

	// Verify customer belongs to this company
	customer, err := s.customerRepo.GetCustomerByID(req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}
	if customer.CompanyID != companyID {
		return nil, errors.New("customer does not belong to this company")
	}

	// Get settings
	settings, err := s.settingsRepo.GetByCompanyID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice settings: %w", err)
	}

	// Assign invoice number
	invoiceNumber, err := s.settingsRepo.GetAndIncrementInvoiceNumber(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to assign invoice number: %w", err)
	}

	// Parse date
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice_date: %w", err)
	}

	paymentTerms := req.PaymentTermsDays
	if paymentTerms <= 0 {
		paymentTerms = settings.DefaultPaymentTermsDays
	}
	dueDate := invoiceDate.AddDate(0, 0, paymentTerms)

	revenueAccount := req.RevenueAccount
	if revenueAccount <= 0 {
		revenueAccount = settings.DefaultRevenueAccount
	}

	// Calculate line totals
	subtotal := decimal.Zero
	vatTotal := decimal.Zero
	var lines []domain.InvoiceLine

	for i, l := range req.Lines {
		if l.Description == "" {
			return nil, fmt.Errorf("line %d: description is required", i+1)
		}
		qty := l.Quantity
		if qty.IsZero() {
			qty = decimal.NewFromInt(1)
		}
		unitPrice := l.UnitPrice
		discountPct := l.DiscountPercent

		// line_total = qty * price * (1 - discount/100)
		lineTotal := qty.Mul(unitPrice)
		if discountPct.IsPositive() {
			discountFactor := decimal.NewFromInt(1).Sub(discountPct.Div(decimal.NewFromInt(100)))
			lineTotal = lineTotal.Mul(discountFactor)
		}
		lineTotal = lineTotal.Round(2)

		// VAT
		vatRate := l.VATRate
		vatAmount := decimal.Zero
		if vatRate > 0 {
			vatAmount = lineTotal.Mul(decimal.NewFromInt(int64(vatRate))).Div(decimal.NewFromInt(100)).Round(2)
		}

		accountNo := l.AccountNo
		if accountNo <= 0 {
			accountNo = revenueAccount
		}

		unit := l.Unit
		if unit == "" {
			unit = "st"
		}

		lines = append(lines, domain.InvoiceLine{
			SortOrder:       i,
			Description:     l.Description,
			Quantity:        qty,
			Unit:            unit,
			UnitPrice:       unitPrice,
			DiscountPercent: discountPct,
			VATRate:         vatRate,
			LineTotal:       lineTotal,
			VATAmount:       vatAmount,
			AccountNo:       accountNo,
		})

		subtotal = subtotal.Add(lineTotal)
		vatTotal = vatTotal.Add(vatAmount)
	}

	// Swedish oresavrundning: round total to nearest whole krona
	rawTotal := subtotal.Add(vatTotal)
	roundedTotal := rawTotal.Round(0)
	rounding := roundedTotal.Sub(rawTotal)

	// Generate OCR
	ocrNumber := GenerateOCRNumber(invoiceNumber)

	invoice := &domain.Invoice{
		CompanyID:        companyID,
		CustomerID:       req.CustomerID,
		InvoiceNumber:    invoiceNumber,
		OCRNumber:        ocrNumber,
		Status:           "draft",
		InvoiceDate:      domain.FlexibleDate{Time: invoiceDate},
		DueDate:          domain.FlexibleDate{Time: dueDate},
		PaymentTermsDays: paymentTerms,
		Currency:         "SEK",
		Subtotal:         subtotal,
		VATTotal:         vatTotal,
		Rounding:         rounding,
		Total:            roundedTotal,
		Notes:            req.Notes,
		YourReference:    req.YourReference,
		OurReference:     req.OurReference,
		Bankgiro:         settings.Bankgiro,
		Plusgiro:         settings.Plusgiro,
		FSkattText:       settings.FSkattText,
		PaymentAccount:   settings.DefaultPaymentAccount,
		RevenueAccount:   revenueAccount,
		CreatedBy:        userID,
	}

	if err := s.invoiceRepo.CreateInvoice(invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create lines
	for i := range lines {
		lines[i].InvoiceID = invoice.InvoiceID
		if err := s.lineRepo.CreateLine(&lines[i]); err != nil {
			return nil, fmt.Errorf("failed to create invoice line: %w", err)
		}
	}

	invoice.Lines = lines
	invoice.Customer = customer
	return invoice, nil
}

// FinalizeInvoice transitions draft → sent and creates the sales voucher
func (s *InvoiceService) FinalizeInvoice(invoiceID, userID int) (*domain.Invoice, error) {
	invoice, err := s.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice.Status != "draft" {
		return nil, errors.New("only draft invoices can be finalized")
	}

	lines, err := s.lineRepo.GetLinesByInvoiceID(invoiceID)
	if err != nil {
		return nil, err
	}

	customer, err := s.customerRepo.GetCustomerByID(invoice.CustomerID)
	if err != nil {
		return nil, err
	}

	// Create the sales voucher
	period := invoice.InvoiceDate.Time.Format("2006-01")
	voucher := &domain.Voucher{
		Date:        invoice.InvoiceDate,
		Description: fmt.Sprintf("Faktura %d - %s", invoice.InvoiceNumber, customer.Name),
		Reference:   fmt.Sprintf("F-%d", invoice.InvoiceNumber),
		TotalAmount: invoice.Total,
		Period:      period,
		CreatedBy:   userID,
		CompanyID:   invoice.CompanyID,
	}

	if err := s.voucherService.CreateVoucher(voucher); err != nil {
		return nil, fmt.Errorf("failed to create sales voucher: %w", err)
	}

	// Debit 1510 Kundfordringar for the total (incl. VAT + rounding)
	err = s.lineItemRepo.CreateLineItem(&domain.LineItem{
		VoucherID:    voucher.VoucherID,
		AccountNo:    AccountKundfordringar,
		DebitAmount:  invoice.Total,
		CreditAmount: decimal.Zero,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create debit line: %w", err)
	}

	// Credit revenue accounts per line (grouped by account)
	revenueByAccount := make(map[int]decimal.Decimal)
	vatByRate := make(map[int]decimal.Decimal)

	for _, line := range lines {
		revenueByAccount[line.AccountNo] = revenueByAccount[line.AccountNo].Add(line.LineTotal)
		if line.VATRate > 0 {
			vatByRate[line.VATRate] = vatByRate[line.VATRate].Add(line.VATAmount)
		}
	}

	for accountNo, amount := range revenueByAccount {
		err = s.lineItemRepo.CreateLineItem(&domain.LineItem{
			VoucherID:    voucher.VoucherID,
			AccountNo:    accountNo,
			DebitAmount:  decimal.Zero,
			CreditAmount: amount,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create revenue line: %w", err)
		}
	}

	// Credit VAT accounts per rate
	for rate, amount := range vatByRate {
		vatAccount := vatAccountForRate(rate)
		if vatAccount == 0 {
			continue
		}
		err = s.lineItemRepo.CreateLineItem(&domain.LineItem{
			VoucherID:    voucher.VoucherID,
			AccountNo:    vatAccount,
			DebitAmount:  decimal.Zero,
			CreditAmount: amount,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create VAT line: %w", err)
		}
	}

	// Rounding adjustment (3740)
	if !invoice.Rounding.IsZero() {
		var debit, credit decimal.Decimal
		if invoice.Rounding.IsPositive() {
			debit = invoice.Rounding
		} else {
			credit = invoice.Rounding.Abs()
		}
		err = s.lineItemRepo.CreateLineItem(&domain.LineItem{
			VoucherID:    voucher.VoucherID,
			AccountNo:    AccountOresavrundning,
			DebitAmount:  debit,
			CreditAmount: credit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create rounding line: %w", err)
		}
	}

	// Update invoice status
	now := time.Now().Format("2006-01-02T15:04:05Z")
	if err := s.invoiceRepo.MarkAsSent(invoiceID, now); err != nil {
		return nil, err
	}
	if err := s.invoiceRepo.LinkSalesVoucher(invoiceID, voucher.VoucherID); err != nil {
		return nil, err
	}

	return s.invoiceRepo.GetInvoiceByID(invoiceID)
}

// MarkAsPaid creates a payment voucher and marks the invoice as paid
func (s *InvoiceService) MarkAsPaid(invoiceID, userID int, paymentDate string, paymentAccountNo int) (*domain.Invoice, error) {
	invoice, err := s.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice.Status != "sent" && invoice.Status != "overdue" {
		return nil, errors.New("only sent or overdue invoices can be marked as paid")
	}

	customer, err := s.customerRepo.GetCustomerByID(invoice.CustomerID)
	if err != nil {
		return nil, err
	}

	pDate, err := time.Parse("2006-01-02", paymentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid payment_date: %w", err)
	}

	if paymentAccountNo <= 0 {
		paymentAccountNo = invoice.PaymentAccount
	}

	period := pDate.Format("2006-01")
	voucher := &domain.Voucher{
		Date:        domain.FlexibleDate{Time: pDate},
		Description: fmt.Sprintf("Betalning faktura %d - %s", invoice.InvoiceNumber, customer.Name),
		Reference:   fmt.Sprintf("F-%d", invoice.InvoiceNumber),
		TotalAmount: invoice.Total,
		Period:      period,
		CreatedBy:   userID,
		CompanyID:   invoice.CompanyID,
	}

	if err := s.voucherService.CreateVoucher(voucher); err != nil {
		return nil, fmt.Errorf("failed to create payment voucher: %w", err)
	}

	// Debit bank/payment account
	err = s.lineItemRepo.CreateLineItem(&domain.LineItem{
		VoucherID:    voucher.VoucherID,
		AccountNo:    paymentAccountNo,
		DebitAmount:  invoice.Total,
		CreditAmount: decimal.Zero,
	})
	if err != nil {
		return nil, err
	}

	// Credit 1510 Kundfordringar
	err = s.lineItemRepo.CreateLineItem(&domain.LineItem{
		VoucherID:    voucher.VoucherID,
		AccountNo:    AccountKundfordringar,
		DebitAmount:  decimal.Zero,
		CreditAmount: invoice.Total,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().Format("2006-01-02T15:04:05Z")
	if err := s.invoiceRepo.MarkAsPaid(invoiceID, invoice.Total, now); err != nil {
		return nil, err
	}
	if err := s.invoiceRepo.LinkPaymentVoucher(invoiceID, voucher.VoucherID); err != nil {
		return nil, err
	}

	return s.invoiceRepo.GetInvoiceByID(invoiceID)
}

// CancelInvoice reverses the sales voucher and marks cancelled
func (s *InvoiceService) CancelInvoice(invoiceID, userID int) (*domain.Invoice, error) {
	invoice, err := s.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice.Status == "cancelled" {
		return nil, errors.New("invoice is already cancelled")
	}
	if invoice.Status == "paid" {
		return nil, errors.New("cannot cancel a paid invoice")
	}

	// If there's a sales voucher, reverse it
	if invoice.SalesVoucherID != nil {
		_, err := s.voucherService.CreateCorrectionVoucher(*invoice.SalesVoucherID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to reverse sales voucher: %w", err)
		}
	}

	now := time.Now().Format("2006-01-02T15:04:05Z")
	if err := s.invoiceRepo.MarkAsCancelled(invoiceID, now); err != nil {
		return nil, err
	}

	return s.invoiceRepo.GetInvoiceByID(invoiceID)
}

// GetInvoiceWithDetails fetches invoice + lines + customer
func (s *InvoiceService) GetInvoiceWithDetails(invoiceID int) (*domain.Invoice, error) {
	invoice, err := s.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return nil, err
	}
	lines, err := s.lineRepo.GetLinesByInvoiceID(invoiceID)
	if err != nil {
		return nil, err
	}
	customer, err := s.customerRepo.GetCustomerByID(invoice.CustomerID)
	if err != nil {
		return nil, err
	}

	invoice.Lines = make([]domain.InvoiceLine, len(lines))
	for i, l := range lines {
		invoice.Lines[i] = *l
	}
	invoice.Customer = customer
	return invoice, nil
}

func (s *InvoiceService) GetInvoicesByCompanyID(companyID int) ([]*domain.Invoice, error) {
	return s.invoiceRepo.GetInvoicesByCompanyID(companyID)
}

func (s *InvoiceService) GetInvoicesByStatus(companyID int, status string) ([]*domain.Invoice, error) {
	return s.invoiceRepo.GetInvoicesByStatus(companyID, status)
}

func (s *InvoiceService) DeleteInvoice(invoiceID int) error {
	invoice, err := s.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return err
	}
	if invoice.Status != "draft" {
		return errors.New("only draft invoices can be deleted")
	}
	if err := s.lineRepo.DeleteLinesByInvoiceID(invoiceID); err != nil {
		return err
	}
	return s.invoiceRepo.DeleteInvoice(invoiceID)
}

// CheckOverdueInvoices marks sent invoices past their due date as overdue
func (s *InvoiceService) CheckOverdueInvoices(companyID int) error {
	overdue, err := s.invoiceRepo.GetOverdueInvoices(companyID)
	if err != nil {
		return err
	}
	for _, inv := range overdue {
		_ = s.invoiceRepo.UpdateInvoiceStatus(inv.InvoiceID, "overdue")
	}
	return nil
}

// GenerateOCRNumber creates a Swedish OCR reference using Luhn check digit
func GenerateOCRNumber(invoiceNumber int) string {
	numStr := strconv.Itoa(invoiceNumber)
	sum := 0
	double := true
	for i := len(numStr) - 1; i >= 0; i-- {
		digit := int(numStr[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	checkDigit := (10 - (sum % 10)) % 10
	return numStr + strconv.Itoa(checkDigit)
}

func vatAccountForRate(rate int) int {
	switch rate {
	case 25:
		return AccountUtgMoms25
	case 12:
		return AccountUtgMoms12
	case 6:
		return AccountUtgMoms6
	default:
		return 0
	}
}
