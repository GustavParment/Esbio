package repository

import (
	"esbio/api/internal/domain"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

type ReportRepository interface {
	GetIncomeStatement(fromDate, toDate string, companyID int) (*domain.IncomeStatement, error)
	GetBalanceSheet(asOfDate string, companyID int) (*domain.BalanceSheet, error)
	GetVATReport(fromDate, toDate string, companyID int) (*domain.VATReport, error)
	GetAccountBalances(beforeDate string, companyID int) ([]domain.AccountBalance, error)
	GetVouchersWithLines(fromDate, toDate string, companyID int) ([]*domain.Voucher, error)
	GetUsedAccounts(companyID int) ([]*domain.Account, error)
}

type reportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetIncomeStatement(fromDate, toDate string, companyID int) (*domain.IncomeStatement, error) {
	query := `
		SELECT
			a.account_no,
			a.account_name,
			a.type,
			SUM(l.debit_amount - l.credit_amount) as balance
		FROM line_items l
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		INNER JOIN accounts a ON l.account_no = a.account_no
		WHERE v.date >= $1
		  AND v.date <= $2
		  AND v.corrected_by_voucher_id IS NULL
		  AND v.company_id = $3
		  AND a.type = 'P&L'
		GROUP BY a.account_no, a.account_name, a.type
		HAVING SUM(l.debit_amount - l.credit_amount) != 0
		ORDER BY a.account_no
	`

	rows, err := r.db.Query(query, fromDate, toDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query income statement: %w", err)
	}
	defer rows.Close()

	statement := &domain.IncomeStatement{
		Income:   make([]domain.IncomeStatementEntry, 0),
		Expenses: make([]domain.IncomeStatementEntry, 0),
	}
	statement.Period.FromDate = fromDate
	statement.Period.ToDate = toDate

	for rows.Next() {
		var accountNo int
		var accountName string
		var accountType string
		var balance decimal.Decimal

		err := rows.Scan(&accountNo, &accountName, &accountType, &balance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Income accounts (3000-3999) have credit balances (negative in debit-credit math).
		// Negate so revenue displays as positive.
		// Expense accounts (4000-8999) have debit balances (positive in debit-credit math).
		// Negate so expenses display as negative.
		displayBalance := balance
		if accountNo >= 3000 && accountNo < 4000 {
			displayBalance = balance.Neg()
		} else if accountNo >= 4000 && accountNo < 9000 {
			displayBalance = balance.Neg()
		}

		entry := domain.IncomeStatementEntry{
			AccountNo:   accountNo,
			AccountName: accountName,
			Balance:     displayBalance,
		}

		if accountNo >= 3000 && accountNo < 4000 {
			statement.Income = append(statement.Income, entry)
			statement.TotalIncome = statement.TotalIncome.Add(displayBalance)
		} else if accountNo >= 4000 && accountNo < 9000 {
			statement.Expenses = append(statement.Expenses, entry)
			statement.TotalExpenses = statement.TotalExpenses.Add(displayBalance)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Net result: positive income + negative expenses = profit (positive) or loss (negative)
	statement.NetResult = statement.TotalIncome.Add(statement.TotalExpenses)

	return statement, nil
}

func (r *reportRepository) GetBalanceSheet(asOfDate string, companyID int) (*domain.BalanceSheet, error) {
	// Get BS account balances (all transactions up to asOfDate)
	bsQuery := `
		SELECT
			a.account_no,
			a.account_name,
			a.account_group,
			SUM(l.debit_amount - l.credit_amount) as balance
		FROM line_items l
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		INNER JOIN accounts a ON l.account_no = a.account_no
		WHERE v.date <= $1
		  AND v.corrected_by_voucher_id IS NULL
		  AND v.company_id = $2
		  AND a.type = 'BS'
		GROUP BY a.account_no, a.account_name, a.account_group
		HAVING SUM(l.debit_amount - l.credit_amount) != 0
		ORDER BY a.account_no
	`

	rows, err := r.db.Query(bsQuery, asOfDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance sheet: %w", err)
	}
	defer rows.Close()

	sheet := &domain.BalanceSheet{
		AsOfDate:          asOfDate,
		Assets:            make([]domain.BalanceSheetEntry, 0),
		EquityLiabilities: make([]domain.BalanceSheetEntry, 0),
	}

	for rows.Next() {
		var accountNo int
		var accountName string
		var accountGroup int
		var balance decimal.Decimal

		if err := rows.Scan(&accountNo, &accountName, &accountGroup, &balance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if accountGroup == 1 {
			// Assets: debit balances are already positive, use as-is
			entry := domain.BalanceSheetEntry{
				AccountNo:   accountNo,
				AccountName: accountName,
				Balance:     balance,
			}
			sheet.Assets = append(sheet.Assets, entry)
			sheet.TotalAssets = sheet.TotalAssets.Add(balance)
		} else if accountGroup == 2 {
			// Equity & Liabilities: credit balances are negative in debit-credit math.
			// Negate so they display as positive.
			negBalance := balance.Neg()
			entry := domain.BalanceSheetEntry{
				AccountNo:   accountNo,
				AccountName: accountName,
				Balance:     negBalance,
			}
			sheet.EquityLiabilities = append(sheet.EquityLiabilities, entry)
			sheet.TotalEquityLiab = sheet.TotalEquityLiab.Add(negBalance)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Calculate P&L net result to include as retained earnings
	plQuery := `
		SELECT
			COALESCE(SUM(l.debit_amount - l.credit_amount), 0) as net_result
		FROM line_items l
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		INNER JOIN accounts a ON l.account_no = a.account_no
		WHERE v.date <= $1
		  AND v.corrected_by_voucher_id IS NULL
		  AND v.company_id = $2
		  AND a.type = 'P&L'
	`

	var netResult decimal.Decimal
	if err := r.db.QueryRow(plQuery, asOfDate, companyID).Scan(&netResult); err != nil {
		return nil, fmt.Errorf("failed to query P&L net result: %w", err)
	}

	// Negate P&L net result: in debit-credit math profit is negative (credit balance).
	// Negate so profit displays as positive on the equity side.
	sheet.NetResult = netResult.Neg()
	sheet.TotalEquityLiab = sheet.TotalEquityLiab.Add(netResult.Neg())

	return sheet, nil
}

func (r *reportRepository) GetVATReport(fromDate, toDate string, companyID int) (*domain.VATReport, error) {
	query := `
		SELECT
			l.tax_code,
			SUM(l.credit_amount - l.debit_amount) as total_sales
		FROM line_items l
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		INNER JOIN accounts a ON l.account_no = a.account_no
		WHERE v.date >= $1
		  AND v.date <= $2
		  AND v.corrected_by_voucher_id IS NULL
		  AND v.company_id = $3
		  AND a.account_no >= 3000 AND a.account_no < 4000
		  AND l.tax_code IS NOT NULL
		GROUP BY l.tax_code
		ORDER BY l.tax_code DESC
	`

	rows, err := r.db.Query(query, fromDate, toDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query VAT report: %w", err)
	}
	defer rows.Close()

	report := &domain.VATReport{
		Entries: make([]domain.VATReportEntry, 0),
	}
	report.Period.FromDate = fromDate
	report.Period.ToDate = toDate

	taxRateLabels := map[int]string{
		25: "25%",
		12: "12%",
		6:  "6%",
		0:  "0%",
	}

	for rows.Next() {
		var taxCode int
		var totalSales decimal.Decimal

		if err := rows.Scan(&taxCode, &totalSales); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		label, ok := taxRateLabels[taxCode]
		if !ok {
			label = fmt.Sprintf("%d%%", taxCode)
		}

		vatAmount := totalSales.Mul(decimal.NewFromInt(int64(taxCode))).Div(decimal.NewFromInt(100))

		entry := domain.VATReportEntry{
			TaxCode:    taxCode,
			TaxRate:    label,
			TotalSales: totalSales,
			TotalVAT:   vatAmount,
		}

		report.Entries = append(report.Entries, entry)
		report.TotalSales = report.TotalSales.Add(totalSales)
		report.TotalVAT = report.TotalVAT.Add(vatAmount)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Fetch input VAT (ingående moms) from accounts 2640-2649
	inputQuery := `
		SELECT
			a.account_no,
			a.account_name,
			SUM(l.debit_amount - l.credit_amount) as amount
		FROM line_items l
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		INNER JOIN accounts a ON l.account_no = a.account_no
		WHERE v.date >= $1
		  AND v.date <= $2
		  AND v.corrected_by_voucher_id IS NULL
		  AND v.company_id = $3
		  AND a.account_no >= 2640 AND a.account_no <= 2649
		GROUP BY a.account_no, a.account_name
		HAVING SUM(l.debit_amount - l.credit_amount) != 0
		ORDER BY a.account_no
	`

	inputRows, err := r.db.Query(inputQuery, fromDate, toDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query input VAT: %w", err)
	}
	defer inputRows.Close()

	report.InputEntries = make([]domain.VATInputEntry, 0)
	for inputRows.Next() {
		var entry domain.VATInputEntry
		if err := inputRows.Scan(&entry.AccountNo, &entry.AccountName, &entry.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan input VAT row: %w", err)
		}
		report.InputEntries = append(report.InputEntries, entry)
		report.TotalInputVAT = report.TotalInputVAT.Add(entry.Amount)
	}

	if err = inputRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating input VAT rows: %w", err)
	}

	// Net VAT = output VAT - input VAT (positive = pay to Skatteverket)
	report.NetVAT = report.TotalVAT.Sub(report.TotalInputVAT)

	return report, nil
}

func (r *reportRepository) GetAccountBalances(beforeDate string, companyID int) ([]domain.AccountBalance, error) {
	query := `
		SELECT a.account_no, a.account_name, a.type,
			   SUM(l.debit_amount - l.credit_amount) as balance
		FROM line_items l
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		INNER JOIN accounts a ON l.account_no = a.account_no
		WHERE v.date < $1
		  AND v.corrected_by_voucher_id IS NULL
		  AND v.company_id = $2
		GROUP BY a.account_no, a.account_name, a.type
		HAVING SUM(l.debit_amount - l.credit_amount) != 0
		ORDER BY a.account_no
	`
	rows, err := r.db.Query(query, beforeDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %w", err)
	}
	defer rows.Close()

	var balances []domain.AccountBalance
	for rows.Next() {
		var b domain.AccountBalance
		if err := rows.Scan(&b.AccountNo, &b.AccountName, &b.Type, &b.Balance); err != nil {
			return nil, fmt.Errorf("failed to scan balance: %w", err)
		}
		balances = append(balances, b)
	}
	return balances, nil
}

func (r *reportRepository) GetVouchersWithLines(fromDate, toDate string, companyID int) ([]*domain.Voucher, error) {
	vQuery := `
		SELECT voucher_id, voucher_number, date, description, reference, total_amount, period,
		       created_by, corrects_voucher_id, corrected_by_voucher_id
		FROM vouchers
		WHERE date >= $1 AND date <= $2
		  AND corrected_by_voucher_id IS NULL
		  AND company_id = $3
		ORDER BY voucher_number
	`
	rows, err := r.db.Query(vQuery, fromDate, toDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vouchers for SIE: %w", err)
	}
	defer rows.Close()

	var vouchers []*domain.Voucher
	for rows.Next() {
		v := &domain.Voucher{}
		if err := rows.Scan(&v.VoucherID, &v.VoucherNumber, &v.Date.Time, &v.Description,
			&v.Reference, &v.TotalAmount, &v.Period, &v.CreatedBy,
			&v.CorrectsVoucherID, &v.CorrectedByVoucherID); err != nil {
			return nil, fmt.Errorf("failed to scan voucher: %w", err)
		}
		vouchers = append(vouchers, v)
	}

	// Fetch line items for each voucher
	liQuery := `
		SELECT line_id, voucher_id, account_no, debit_amount, credit_amount,
		       COALESCE(tax_code, 0), COALESCE(project_id, 0), COALESCE(cost_center_id, 0)
		FROM line_items WHERE voucher_id = $1 ORDER BY line_id
	`
	for _, v := range vouchers {
		liRows, err := r.db.Query(liQuery, v.VoucherID)
		if err != nil {
			return nil, fmt.Errorf("failed to get line items: %w", err)
		}
		for liRows.Next() {
			li := domain.LineItem{}
			if err := liRows.Scan(&li.LineID, &li.VoucherID, &li.AccountNo,
				&li.DebitAmount, &li.CreditAmount, &li.TaxCode, &li.ProjectID, &li.CostCenterID); err != nil {
				liRows.Close()
				return nil, fmt.Errorf("failed to scan line item: %w", err)
			}
			v.Lines = append(v.Lines, li)
		}
		liRows.Close()
	}

	return vouchers, nil
}

func (r *reportRepository) GetUsedAccounts(companyID int) ([]*domain.Account, error) {
	query := `
		SELECT DISTINCT a.account_no, a.account_name, a.account_group, a.tax_standard, a.type, a.standard_side
		FROM accounts a
		INNER JOIN line_items l ON a.account_no = l.account_no
		INNER JOIN vouchers v ON l.voucher_id = v.voucher_id
		WHERE v.company_id = $1
		ORDER BY a.account_no
	`
	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get used accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		a := &domain.Account{}
		if err := rows.Scan(&a.AccountNo, &a.AccountName, &a.AccountGroup,
			&a.TaxStandard, &a.Type, &a.StandardSide); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}
