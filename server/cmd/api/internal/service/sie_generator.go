package service

import (
	"bytes"
	"esbio/api/internal/repository"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

type SIEGenerator struct {
	reportRepo     repository.ReportRepository
	companyService *CompanyService
}

func NewSIEGenerator(reportRepo repository.ReportRepository, companyService *CompanyService) *SIEGenerator {
	return &SIEGenerator{
		reportRepo:     reportRepo,
		companyService: companyService,
	}
}

func (g *SIEGenerator) GenerateSIE4(fromDate, toDate string, companyID int) ([]byte, error) {
	// 1. Fetch company info
	company, err := g.companyService.GetCompanyByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	companyName := company.CompanyName

	// 2. Fetch used accounts
	accounts, err := g.reportRepo.GetUsedAccounts(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	// 3. Fetch opening balances (before fromDate)
	openingBalances, err := g.reportRepo.GetAccountBalances(fromDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get opening balances: %w", err)
	}

	// 4. Fetch vouchers with line items
	vouchers, err := g.reportRepo.GetVouchersWithLines(fromDate, toDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vouchers: %w", err)
	}

	// 5. Calculate closing balances and results
	// Build opening balance map
	ibMap := make(map[int]float64)
	ibTypeMap := make(map[int]string)
	for _, b := range openingBalances {
		ibMap[b.AccountNo] = b.Balance
		ibTypeMap[b.AccountNo] = b.Type
	}

	// Calculate period movements per account
	periodMovement := make(map[int]float64)
	periodType := make(map[int]string)
	for _, v := range vouchers {
		for _, li := range v.Lines {
			movement := li.DebitAmount - li.CreditAmount
			periodMovement[li.AccountNo] += movement
		}
	}

	// Get types for accounts with period movement
	for _, a := range accounts {
		periodType[a.AccountNo] = a.Type
		if _, exists := ibTypeMap[a.AccountNo]; !exists {
			ibTypeMap[a.AccountNo] = a.Type
		}
	}

	// 6. Generate SIE4 content
	var buf bytes.Buffer
	w := func(format string, args ...interface{}) {
		buf.WriteString(fmt.Sprintf(format, args...))
		buf.WriteString("\r\n")
	}

	now := time.Now()
	fromCompact := strings.ReplaceAll(fromDate, "-", "")
	toCompact := strings.ReplaceAll(toDate, "-", "")

	// Header
	w("#FLAGGA 0")
	w("#FORMAT PC8")
	w("#SIETYP 4")
	w("#PROGRAM \"Esbio\" \"1.0\"")
	w("#GEN %s", now.Format("20060102"))
	w("#FNAMN \"%s\"", companyName)
	if company.OrgNumber != "" {
		w("#ORGNR %s", company.OrgNumber)
	}
	w("#KPTYP BAS2024")
	w("#RAR 0 %s %s", fromCompact, toCompact)
	w("")

	// Accounts
	for _, a := range accounts {
		w("#KONTO %d \"%s\"", a.AccountNo, a.AccountName)
	}
	w("")

	// Opening balances (IB) — BS accounts only
	for _, b := range openingBalances {
		if b.Type == "BS" {
			w("#IB 0 %d %.2f", b.AccountNo, b.Balance)
		}
	}
	w("")

	// Closing balances (UB) — BS accounts only
	for accNo, ib := range ibMap {
		if ibTypeMap[accNo] == "BS" {
			ub := ib + periodMovement[accNo]
			if ub != 0 {
				w("#UB 0 %d %.2f", accNo, ub)
			}
		}
	}
	// Also accounts that only have period movement (no IB)
	for accNo, movement := range periodMovement {
		if periodType[accNo] == "BS" {
			if _, hasIB := ibMap[accNo]; !hasIB && movement != 0 {
				w("#UB 0 %d %.2f", accNo, movement)
			}
		}
	}
	w("")

	// Result balances (RES) — P&L accounts only
	for accNo, movement := range periodMovement {
		if periodType[accNo] == "P&L" && movement != 0 {
			w("#RES 0 %d %.2f", accNo, movement)
		}
	}
	w("")

	// Vouchers
	for _, v := range vouchers {
		dateCompact := v.Date.Time.Format("20060102")
		description := strings.ReplaceAll(v.Description, "\"", "'")
		w("#VER \"A\" %d %s \"%s\"", v.VoucherNumber, dateCompact, description)
		w("{")
		for _, li := range v.Lines {
			amount := li.DebitAmount - li.CreditAmount
			w("  #TRANS %d {} %.2f", li.AccountNo, amount)
		}
		w("}")
	}

	// 7. Encode to ISO 8859-1
	encoder := charmap.ISO8859_1.NewEncoder()
	encoded, err := encoder.Bytes(buf.Bytes())
	if err != nil {
		// Fallback to UTF-8 if encoding fails
		return buf.Bytes(), nil
	}

	return encoded, nil
}
