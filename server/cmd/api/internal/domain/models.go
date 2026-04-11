package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
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
    UserID       int    `json:"user_id" validate:"omitempty,gt=0"`
    FirstName    string `json:"first_name" validate:"required,min=1,max=50"`
    LastName     string `json:"last_name" validate:"required,min=1,max=50"`
    Name         string `json:"name"`
    CompanyName  string `json:"company_name"`
    OrgNumber    string `json:"org_number"`
    Email        string `json:"email" validate:"required,email"`
    PasswordHash string `json:"password_hash" validate:"required,min=8"`
    Role         string `json:"role" validate:"omitempty,oneof=Admin Bookkeeper Manager"`
}

type Company struct {
    CompanyID            int     `json:"company_id"`
    CompanyName          string  `json:"company_name"`
    OrgNumber            string  `json:"org_number"`
    Plan                 string  `json:"plan"`
    CreatedBy            int     `json:"created_by"`
    CreatedAt            string  `json:"created_at"`
    UpdatedAt            string  `json:"updated_at"`
    StripeCustomerID     *string `json:"stripe_customer_id,omitempty"`
    StripeSubscriptionID *string `json:"stripe_subscription_id,omitempty"`
    PlanStatus           string  `json:"plan_status"`
}

type AccountBalance struct {
    AccountNo   int             `json:"account_no"`
    AccountName string          `json:"account_name"`
    Type        string          `json:"type"`
    Balance     decimal.Decimal `json:"balance"`
}

type AgentUsage struct {
    UsageID          int    `json:"usage_id"`
    UserID           int    `json:"user_id"`
    CompanyID        int    `json:"company_id"`
    PromptTokens     int    `json:"prompt_tokens"`
    CompletionTokens int    `json:"completion_tokens"`
    TotalTokens      int    `json:"total_tokens"`
    CreatedAt        string `json:"created_at"`
}

type AgentUsageSummary struct {
    TotalPromptTokens     int     `json:"total_prompt_tokens"`
    TotalCompletionTokens int     `json:"total_completion_tokens"`
    TotalTokens           int     `json:"total_tokens"`
    RequestCount          int     `json:"request_count"`
    EstimatedCostSEK      float64 `json:"estimated_cost_sek"`
    MonthPromptTokens     int     `json:"month_prompt_tokens"`
    MonthCompletionTokens int     `json:"month_completion_tokens"`
    MonthTotalTokens      int     `json:"month_total_tokens"`
    MonthRequestCount     int     `json:"month_request_count"`
    MonthEstimatedCostSEK float64 `json:"month_estimated_cost_sek"`
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
    LineID       int             `json:"line_id"`         // Unikt ID för raden
    VoucherID    int             `json:"voucher_id"`      // Foreign Key till VoucherID
    AccountNo    int             `json:"account_no"`      // Foreign Key till Account
    DebitAmount  decimal.Decimal `json:"debit_amount"`    // Belopp i Debet
    CreditAmount decimal.Decimal `json:"credit_amount"`   // Belopp i Kredit
    TaxCode      int             `json:"tax_code"`        // Momskod (t.ex. 25, 12, 6, 0)
    ProjectID    int             `json:"project_id"`      // Valfri: För projektspårning
    CostCenterID int             `json:"cost_center_id"`  // Valfri: För kostnadsställe
}

type Voucher struct {
    VoucherID            int             `json:"voucher_id"`
    VoucherNumber        int             `json:"voucher_number"`
    Date                 FlexibleDate    `json:"date"`
    Description          string          `json:"description"`
    Reference            string          `json:"reference"`
    TotalAmount          decimal.Decimal `json:"total_amount"`
    Period               string          `json:"period"`
    CreatedBy            int             `json:"created_by"`
    CompanyID            int             `json:"company_id"`
    CorrectsVoucherID    *int            `json:"corrects_voucher_id"`     // ID för verifikat som detta rättar
    CorrectedByVoucherID *int            `json:"corrected_by_voucher_id"` // ID för verifikat som rättat detta
    Lines                []LineItem      `json:"lines"`                   // Lista över Verifikatraderna
}

type LedgerEntry struct {
    Date          FlexibleDate    `json:"date"`           // Transaction date
    VoucherID     int             `json:"voucher_id"`     // Voucher ID
    VoucherNumber int             `json:"voucher_number"` // Voucher number (#1, #2, etc.)
    Description   string          `json:"description"`    // Transaction description
    Reference     string          `json:"reference"`      // Invoice/reference number
    DebitAmount   decimal.Decimal `json:"debit_amount"`   // Debit amount
    CreditAmount  decimal.Decimal `json:"credit_amount"`  // Credit amount
    Balance       decimal.Decimal `json:"balance"`        // Running balance (calculated)
}

type IncomeStatementEntry struct {
    AccountNo   int             `json:"account_no"`   // Account number
    AccountName string          `json:"account_name"` // Account name
    Balance     decimal.Decimal `json:"balance"`      // Account balance for period
}

type IncomeStatement struct {
    Period struct {
        FromDate string `json:"from_date"` // Start date (YYYY-MM-DD)
        ToDate   string `json:"to_date"`   // End date (YYYY-MM-DD)
    } `json:"period"`
    Income        []IncomeStatementEntry `json:"income"`         // Revenue accounts (3000-3999)
    Expenses      []IncomeStatementEntry `json:"expenses"`       // Expense accounts (4000-8999)
    TotalIncome   decimal.Decimal        `json:"total_income"`   // Sum of all income
    TotalExpenses decimal.Decimal        `json:"total_expenses"` // Sum of all expenses
    NetResult     decimal.Decimal        `json:"net_result"`     // Total income - Total expenses
}

type BalanceSheetEntry struct {
    AccountNo   int             `json:"account_no"`
    AccountName string          `json:"account_name"`
    Balance     decimal.Decimal `json:"balance"`
}

type BalanceSheet struct {
    AsOfDate          string              `json:"as_of_date"`
    Assets            []BalanceSheetEntry `json:"assets"`             // Group 1: Tillgångar
    EquityLiabilities []BalanceSheetEntry `json:"equity_liabilities"` // Group 2: Eget kapital & skulder
    TotalAssets       decimal.Decimal     `json:"total_assets"`
    TotalEquityLiab   decimal.Decimal     `json:"total_equity_liabilities"`
    NetResult         decimal.Decimal     `json:"net_result"` // P&L result carried into equity
}

type VATReportEntry struct {
    TaxCode      int             `json:"tax_code"`       // 25, 12, 6, 0
    TaxRate      string          `json:"tax_rate"`       // "25%", "12%", "6%", "0%"
    TotalSales   decimal.Decimal `json:"total_sales"`    // Revenue (credit side of income accounts)
    TotalVAT     decimal.Decimal `json:"total_vat"`      // Calculated VAT amount
}

type VATInputEntry struct {
    AccountNo   int             `json:"account_no"`     // 2641, 2642, 2643, etc.
    AccountName string          `json:"account_name"`
    Amount      decimal.Decimal `json:"amount"`         // Input VAT amount (debit balance)
}

type VATReport struct {
    Period struct {
        FromDate string `json:"from_date"`
        ToDate   string `json:"to_date"`
    } `json:"period"`
    Entries       []VATReportEntry `json:"entries"`
    TotalSales    decimal.Decimal  `json:"total_sales"`
    TotalVAT      decimal.Decimal  `json:"total_vat"`
    InputEntries  []VATInputEntry  `json:"input_entries"`
    TotalInputVAT decimal.Decimal  `json:"total_input_vat"`
    NetVAT        decimal.Decimal  `json:"net_vat"`
}

type ScheduledTask struct {
    TaskID             int     `json:"task_id"`
    UserID             int     `json:"user_id"`
    CompanyID          int     `json:"company_id"`
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
    CompanyID      int    `json:"company_id"`
    ConversationID string `json:"conversation_id"`
    Role           string `json:"role"`
    Content        string `json:"content"`
    CreatedAt      string `json:"created_at"`
}

// Bank / Tink Open Banking models

type BankConnection struct {
    ConnectionID        int     `json:"connection_id"`
    CompanyID           int     `json:"company_id"`
    CreatedBy           int     `json:"created_by"`
    TinkUserID          string  `json:"tink_user_id"`
    TinkCredentialsID   *string `json:"tink_credentials_id"`
    BankName            string  `json:"bank_name"`
    Status              string  `json:"status"`
    AccessTokenEncrypted *string `json:"-"`
    ConsentExpiresAt    string  `json:"consent_expires_at"`
    LastSyncedAt        *string `json:"last_synced_at"`
    CreatedAt           string  `json:"created_at"`
    UpdatedAt           string  `json:"updated_at"`
}

type BankAccount struct {
    BankAccountID    int              `json:"bank_account_id"`
    ConnectionID     int              `json:"connection_id"`
    CompanyID        int              `json:"company_id"`
    TinkAccountID    string           `json:"tink_account_id"`
    AccountName      string           `json:"account_name"`
    IBAN             *string          `json:"iban"`
    AccountNumber    *string          `json:"account_number"`
    Currency         string           `json:"currency"`
    BalanceAmount    *decimal.Decimal `json:"balance_amount"`
    BalanceUpdatedAt *string          `json:"balance_updated_at"`
    MappedAccountNo  *int             `json:"mapped_account_no"`
    CreatedAt        string           `json:"created_at"`
    UpdatedAt        string           `json:"updated_at"`
}

type BankTransaction struct {
    BankTransactionID    int             `json:"bank_transaction_id"`
    BankAccountID        int             `json:"bank_account_id"`
    CompanyID            int             `json:"company_id"`
    TinkTransactionID    string          `json:"tink_transaction_id"`
    BookedDate           string          `json:"booked_date"`
    Amount               decimal.Decimal `json:"amount"`
    Currency             string          `json:"currency"`
    DescriptionOriginal  string          `json:"description_original"`
    DescriptionDisplay   *string         `json:"description_display"`
    MerchantName         *string         `json:"merchant_name"`
    MerchantCategoryCode *string         `json:"merchant_category_code"`
    TinkCategoryID       *string         `json:"tink_category_id"`
    Status               string          `json:"status"`
    ImportStatus         string          `json:"import_status"`
    SuggestedAccountNo   *int            `json:"suggested_account_no"`
    SuggestedDescription *string         `json:"suggested_description"`
    VoucherID            *int            `json:"voucher_id"`
    CreatedAt            string          `json:"created_at"`
    UpdatedAt            string          `json:"updated_at"`
}

// Receipt scanning types
type ReceiptScanResult struct {
	Date           string            `json:"date"`
	Vendor         string            `json:"vendor"`
	Description    string            `json:"description"`
	TotalAmount    decimal.Decimal   `json:"total_amount"`
	VATRate        int               `json:"vat_rate"`
	VATAmount      decimal.Decimal   `json:"vat_amount"`
	AmountExclVAT  decimal.Decimal   `json:"amount_excl_vat"`
	Currency       string            `json:"currency"`
	SuggestedLines []ReceiptLineItem `json:"suggested_lines"`
}

type ReceiptLineItem struct {
	AccountNo    int             `json:"account_no"`
	AccountName  string          `json:"account_name"`
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`
	TaxCode      int             `json:"tax_code"`
}

type CategorizationRule struct {
    RuleID              int     `json:"rule_id"`
    CompanyID           int     `json:"company_id"`
    MatchPattern        string  `json:"match_pattern"`
    MatchType           string  `json:"match_type"`
    AccountNo           int     `json:"account_no"`
    DescriptionTemplate *string `json:"description_template"`
    TaxCode             *int    `json:"tax_code"`
    Confidence          float64 `json:"confidence"`
    UseCount            int     `json:"use_count"`
    CreatedAt           string  `json:"created_at"`
    UpdatedAt           string  `json:"updated_at"`
}