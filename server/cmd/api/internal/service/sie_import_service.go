package service

import (
	"esbio/api/internal/domain"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type SIEImportService struct {
	voucherService  *VoucherService
	accountService  *AccountService
	lineItemService *LineItemService
}

func NewSIEImportService(voucherService *VoucherService, accountService *AccountService, lineItemService *LineItemService) *SIEImportService {
	return &SIEImportService{
		voucherService:  voucherService,
		accountService:  accountService,
		lineItemService: lineItemService,
	}
}

// SIEImportSummary is returned by both Preview and Import.
type SIEImportSummary struct {
	CompanyName     string             `json:"company_name"`
	OrgNumber       string             `json:"org_number"`
	ChartType       string             `json:"chart_type"`
	FiscalFrom      string             `json:"fiscal_from"`
	FiscalTo        string             `json:"fiscal_to"`
	Encoding        string             `json:"encoding"`
	AccountsInFile  int                `json:"accounts_in_file"`
	AccountsToAdd   int                `json:"accounts_to_add"`
	VouchersInFile  int                `json:"vouchers_in_file"`
	UnbalancedCount int                `json:"unbalanced_count"`
	Warnings        []string           `json:"warnings"`
	DryRun          bool               `json:"dry_run"`
	Imported        *SIEImportResult   `json:"imported,omitempty"`
}

type SIEImportResult struct {
	AccountsCreated int `json:"accounts_created"`
	AccountsSkipped int `json:"accounts_skipped"`
	VouchersCreated int `json:"vouchers_created"`
	VouchersSkipped int `json:"vouchers_skipped"`
}

// PreviewImport parses the file and reports what would happen, without writing.
func (s *SIEImportService) PreviewImport(raw []byte) (*SIEImportSummary, error) {
	parsed, err := ParseSIE(raw)
	if err != nil {
		return nil, err
	}
	summary := s.buildSummary(parsed, true)
	return summary, nil
}

// Import parses the file and writes accounts + vouchers to the company.
func (s *SIEImportService) Import(raw []byte, companyID, userID int) (*SIEImportSummary, error) {
	if companyID <= 0 {
		return nil, errors.New("invalid company ID")
	}
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	parsed, err := ParseSIE(raw)
	if err != nil {
		return nil, err
	}

	summary := s.buildSummary(parsed, false)
	result := &SIEImportResult{}

	// 1. Create missing accounts
	for _, a := range parsed.Accounts {
		existing, _ := s.accountService.GetAccountByNo(a.AccountNo)
		if existing != nil {
			result.AccountsSkipped++
			continue
		}
		acc := buildAccountFromBAS(a.AccountNo, a.AccountName)
		if err := s.accountService.CreateAccount(acc); err != nil {
			summary.Warnings = append(summary.Warnings,
				fmt.Sprintf("could not create account %d: %v", a.AccountNo, err))
			continue
		}
		result.AccountsCreated++
	}

	// 2. Create vouchers
	for _, v := range parsed.Vouchers {
		// Skip unbalanced vouchers — never persist invalid bookkeeping
		sum := decimal.Zero
		for _, l := range v.Lines {
			sum = sum.Add(l.Amount)
		}
		if !sum.IsZero() {
			result.VouchersSkipped++
			continue
		}
		if len(v.Lines) == 0 {
			result.VouchersSkipped++
			continue
		}

		voucher := &domain.Voucher{
			Date:        domain.FlexibleDate{Time: v.Date},
			Description: defaultIfBlank(v.Description, fmt.Sprintf("SIE-import %s/%d", v.Series, v.Number)),
			Reference:   fmt.Sprintf("SIE %s/%d", v.Series, v.Number),
			Period:      v.Date.Format("2006-01"),
			TotalAmount: voucherTotal(v.Lines),
			CreatedBy:   userID,
			CompanyID:   companyID,
		}

		// Build line items: SIE amount > 0 → debit, < 0 → credit
		voucher.Lines = make([]domain.LineItem, 0, len(v.Lines))
		for _, l := range v.Lines {
			li := domain.LineItem{AccountNo: l.AccountNo}
			if l.Amount.IsNegative() {
				li.CreditAmount = l.Amount.Abs()
				li.DebitAmount = decimal.Zero
			} else {
				li.DebitAmount = l.Amount
				li.CreditAmount = decimal.Zero
			}
			voucher.Lines = append(voucher.Lines, li)
		}

		if err := s.voucherService.CreateVoucher(voucher); err != nil {
			summary.Warnings = append(summary.Warnings,
				fmt.Sprintf("could not create voucher %s/%d: %v", v.Series, v.Number, err))
			result.VouchersSkipped++
			continue
		}

		// Persist line items now that the voucher has an ID
		lineErr := false
		for i := range voucher.Lines {
			line := voucher.Lines[i]
			line.VoucherID = voucher.VoucherID
			if line.DebitAmount.IsZero() && line.CreditAmount.IsZero() {
				continue // skip zero-amount artifacts
			}
			if err := s.lineItemService.CreateLineItem(&line); err != nil {
				summary.Warnings = append(summary.Warnings,
					fmt.Sprintf("voucher %s/%d line on account %d failed: %v",
						v.Series, v.Number, line.AccountNo, err))
				lineErr = true
			}
		}
		if lineErr {
			summary.Warnings = append(summary.Warnings,
				fmt.Sprintf("voucher %s/%d created with one or more failed lines — review manually", v.Series, v.Number))
		}
		result.VouchersCreated++
	}

	summary.Imported = result
	return summary, nil
}

func (s *SIEImportService) buildSummary(p *ParsedSIE, dryRun bool) *SIEImportSummary {
	unbalanced := 0
	for _, v := range p.Vouchers {
		sum := decimal.Zero
		for _, l := range v.Lines {
			sum = sum.Add(l.Amount)
		}
		if !sum.IsZero() {
			unbalanced++
		}
	}

	// Count accounts that don't exist yet
	toAdd := 0
	for _, a := range p.Accounts {
		existing, _ := s.accountService.GetAccountByNo(a.AccountNo)
		if existing == nil {
			toAdd++
		}
	}

	return &SIEImportSummary{
		CompanyName:     p.CompanyName,
		OrgNumber:       p.OrgNumber,
		ChartType:       p.ChartType,
		FiscalFrom:      p.FiscalFrom,
		FiscalTo:        p.FiscalTo,
		Encoding:        p.Encoding,
		AccountsInFile:  len(p.Accounts),
		AccountsToAdd:   toAdd,
		VouchersInFile:  len(p.Vouchers),
		UnbalancedCount: unbalanced,
		Warnings:        p.Warnings,
		DryRun:          dryRun,
	}
}

// buildAccountFromBAS infers BAS classification from the account number.
// Group = first digit; type and standard side follow BAS conventions.
func buildAccountFromBAS(accountNo int, name string) *domain.Account {
	group := accountNo / 1000
	if group < 1 {
		group = 1
	}
	if group > 8 {
		group = 8
	}

	accountType := "P&L"
	if group <= 2 {
		accountType = "BS"
	}

	standardSide := "Debit"
	switch group {
	case 2, 3:
		standardSide = "Credit"
	}

	return &domain.Account{
		AccountNo:    accountNo,
		AccountName:  name,
		AccountGroup: group,
		Type:         accountType,
		StandardSide: standardSide,
	}
}

func voucherTotal(lines []ParsedTrans) decimal.Decimal {
	total := decimal.Zero
	for _, l := range lines {
		if l.Amount.IsPositive() {
			total = total.Add(l.Amount)
		}
	}
	return total
}

func defaultIfBlank(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
