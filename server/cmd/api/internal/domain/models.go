package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// FlexibleDate handles both date-only (YYYY-MM-DD) and full timestamp formats
type FlexibleDate struct {
	time.Time
}

func (fd *FlexibleDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	// Try date-only format first (YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		fd.Time = t
		return nil
	}

	// Try RFC3339 format
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		fd.Time = t
		return nil
	}

	return fmt.Errorf("could not parse date: %s", s)
}

func (fd FlexibleDate) MarshalJSON() ([]byte, error) {
	// Always output in RFC3339 format
	return json.Marshal(fd.Time)
}

type User struct {
    UserID      int    `json:"user_id" validate:"omitempty,gt=0"`     // Unikt ID
    Name        string `json:"name" validate:"required,min=2,max=100"`        // Användarens fullständiga namn
    Email       string `json:"email" validate:"required,email"`       // E-post (unikt)
    PasswordHash string `json:"password_hash" validate:"required,min=8"` // Krypterat lösenord
    Role        string `json:"role" validate:"omitempty,oneof=Admin Bookkeeper Manager"`        // T.ex. "Admin", "Bookkeeper"
}

type Account struct {
    AccountNo   int    `json:"account_no"`    // Kontonummer (Primary Key)
    AccountName string `json:"account_name"`  // Kontots namn
    AccountGroup int   `json:"account_group"` // Kontogrupp (1-8 i BAS-planen)
    TaxStandard string `json:"tax_standard"`  // Standardmomskod (t.ex. "25%")
    Type        string `json:"type"`          // "P&L" (Resultat) eller "BS" (Balans)
    StandardSide string `json:"standard_side"` // "Debit" eller "Credit"
}

type LineItem struct {
    LineID       int     `json:"line_id"`         // Unikt ID för raden
    VoucherID    int     `json:"voucher_id"`      // Foreign Key till VoucherID
    AccountNo    int     `json:"account_no"`      // Foreign Key till Account
    DebitAmount  float64 `json:"debit_amount"`    // Belopp i Debet
    CreditAmount float64 `json:"credit_amount"`   // Belopp i Kredit
    TaxCode      int     `json:"tax_code"`        // Momskod (t.ex. 25, 12, 6, 0)
    ProjectID    int     `json:"project_id"`      // Valfri: För projektspårning
    CostCenterID int     `json:"cost_center_id"`  // Valfri: För kostnadsställe
}

type Voucher struct {
    VoucherID            int          `json:"voucher_id"`              // Unikt ID
    VoucherNumber        int          `json:"voucher_number"`          // Löpnummer för verifikat (1, 2, 3...)
    Date                 FlexibleDate `json:"date"`                    // Datum då händelsen inträffade
    Description          string       `json:"description"`             // Beskrivning av transaktionen
    Reference            string       `json:"reference"`               // Fakturanummer, kvitto-ID, etc.
    TotalAmount          float64      `json:"total_amount"`            // Totalbelopp
    Period               string       `json:"period"`                  // Period (t.ex. "2025-01")
    CreatedBy            int          `json:"created_by"`              // Foreign Key till UserID
    CorrectsVoucherID    *int         `json:"corrects_voucher_id"`     // ID för verifikat som detta rättar
    CorrectedByVoucherID *int         `json:"corrected_by_voucher_id"` // ID för verifikat som rättat detta
    Lines                []LineItem   `json:"lines"`                   // Lista över Verifikatraderna
}

type LedgerEntry struct {
    Date          FlexibleDate `json:"date"`           // Transaction date
    VoucherID     int          `json:"voucher_id"`     // Voucher ID
    VoucherNumber int          `json:"voucher_number"` // Voucher number (#1, #2, etc.)
    Description   string       `json:"description"`    // Transaction description
    Reference     string       `json:"reference"`      // Invoice/reference number
    DebitAmount   float64      `json:"debit_amount"`   // Debit amount
    CreditAmount  float64      `json:"credit_amount"`  // Credit amount
    Balance       float64      `json:"balance"`        // Running balance (calculated)
}

type IncomeStatementEntry struct {
    AccountNo   int     `json:"account_no"`   // Account number
    AccountName string  `json:"account_name"` // Account name
    Balance     float64 `json:"balance"`      // Account balance for period
}

type IncomeStatement struct {
    Period struct {
        FromDate string `json:"from_date"` // Start date (YYYY-MM-DD)
        ToDate   string `json:"to_date"`   // End date (YYYY-MM-DD)
    } `json:"period"`
    Income        []IncomeStatementEntry `json:"income"`         // Revenue accounts (3000-3999)
    Expenses      []IncomeStatementEntry `json:"expenses"`       // Expense accounts (4000-8999)
    TotalIncome   float64                `json:"total_income"`   // Sum of all income
    TotalExpenses float64                `json:"total_expenses"` // Sum of all expenses
    NetResult     float64                `json:"net_result"`     // Total income - Total expenses
}

type BalanceSheetEntry struct {
    AccountNo   int     `json:"account_no"`
    AccountName string  `json:"account_name"`
    Balance     float64 `json:"balance"`
}

type BalanceSheet struct {
    AsOfDate          string              `json:"as_of_date"`
    Assets            []BalanceSheetEntry `json:"assets"`             // Group 1: Tillgångar
    EquityLiabilities []BalanceSheetEntry `json:"equity_liabilities"` // Group 2: Eget kapital & skulder
    TotalAssets       float64             `json:"total_assets"`
    TotalEquityLiab   float64             `json:"total_equity_liabilities"`
    NetResult         float64             `json:"net_result"` // P&L result carried into equity
}

type VATReportEntry struct {
    TaxCode      int     `json:"tax_code"`       // 25, 12, 6, 0
    TaxRate      string  `json:"tax_rate"`       // "25%", "12%", "6%", "0%"
    TotalSales   float64 `json:"total_sales"`    // Revenue (credit side of income accounts)
    TotalVAT     float64 `json:"total_vat"`      // Calculated VAT amount
}

type VATReport struct {
    Period struct {
        FromDate string `json:"from_date"`
        ToDate   string `json:"to_date"`
    } `json:"period"`
    Entries       []VATReportEntry `json:"entries"`
    TotalSales    float64          `json:"total_sales"`
    TotalVAT      float64          `json:"total_vat"`
}

type ScheduledTask struct {
    TaskID             int     `json:"task_id"`
    UserID             int     `json:"user_id"`
    Description        string  `json:"description"`
    Prompt             string  `json:"prompt"`
    TemplateVoucherID  *int    `json:"template_voucher_id"`
    DayOfMonth         int     `json:"day_of_month"`
    Active             bool    `json:"active"`
    LastRunAt          *string `json:"last_run_at"`
    NextRunAt          *string `json:"next_run_at"`
    CreatedAt          string  `json:"created_at"`
    UpdatedAt          string  `json:"updated_at"`
}

type AgentMessage struct {
    MessageID      int    `json:"message_id"`
    UserID         int    `json:"user_id"`
    ConversationID string `json:"conversation_id"`
    Role           string `json:"role"`
    Content        string `json:"content"`
    CreatedAt      string `json:"created_at"`
}