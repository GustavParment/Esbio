package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
)

type ReportRepository interface {
	GetIncomeStatement(fromDate, toDate string) (*domain.IncomeStatement, error)
	GetBalanceSheet(asOfDate string) (*domain.BalanceSheet, error)
	GetVATReport(fromDate, toDate string) (*domain.VATReport, error)
}

type reportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetIncomeStatement(fromDate, toDate string) (*domain.IncomeStatement, error) {
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
		  AND a.type = 'P&L'
		GROUP BY a.account_no, a.account_name, a.type
		HAVING SUM(l.debit_amount - l.credit_amount) != 0
		ORDER BY a.account_no
	`

	rows, err := r.db.Query(query, fromDate, toDate)
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
		var balance float64

		err := rows.Scan(&accountNo, &accountName, &accountType, &balance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		entry := domain.IncomeStatementEntry{
			AccountNo:   accountNo,
			AccountName: accountName,
			Balance:     balance,
		}

		// Income accounts (3000-3999): positive balance means income
		// Expense accounts (4000-8999): negative balance means expense
		if accountNo >= 3000 && accountNo < 4000 {
			statement.Income = append(statement.Income, entry)
			statement.TotalIncome += balance
		} else if accountNo >= 4000 && accountNo < 9000 {
			statement.Expenses = append(statement.Expenses, entry)
			statement.TotalExpenses += balance
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Calculate net result
	statement.NetResult = statement.TotalIncome + statement.TotalExpenses

	return statement, nil
}

func (r *reportRepository) GetBalanceSheet(asOfDate string) (*domain.BalanceSheet, error) {
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
		  AND a.type = 'BS'
		GROUP BY a.account_no, a.account_name, a.account_group
		HAVING SUM(l.debit_amount - l.credit_amount) != 0
		ORDER BY a.account_no
	`

	rows, err := r.db.Query(bsQuery, asOfDate)
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
		var balance float64

		if err := rows.Scan(&accountNo, &accountName, &accountGroup, &balance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		entry := domain.BalanceSheetEntry{
			AccountNo:   accountNo,
			AccountName: accountName,
			Balance:     balance,
		}

		if accountGroup == 1 {
			sheet.Assets = append(sheet.Assets, entry)
			sheet.TotalAssets += balance
		} else if accountGroup == 2 {
			sheet.EquityLiabilities = append(sheet.EquityLiabilities, entry)
			sheet.TotalEquityLiab += balance
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
		  AND a.type = 'P&L'
	`

	var netResult float64
	if err := r.db.QueryRow(plQuery, asOfDate).Scan(&netResult); err != nil {
		return nil, fmt.Errorf("failed to query P&L net result: %w", err)
	}

	sheet.NetResult = netResult
	// Net result is added to equity side (credit = negative in debit-credit math)
	sheet.TotalEquityLiab += netResult

	return sheet, nil
}

func (r *reportRepository) GetVATReport(fromDate, toDate string) (*domain.VATReport, error) {
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
		  AND a.account_no >= 3000 AND a.account_no < 4000
		  AND l.tax_code IS NOT NULL
		GROUP BY l.tax_code
		ORDER BY l.tax_code DESC
	`

	rows, err := r.db.Query(query, fromDate, toDate)
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
		var totalSales float64

		if err := rows.Scan(&taxCode, &totalSales); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		label, ok := taxRateLabels[taxCode]
		if !ok {
			label = fmt.Sprintf("%d%%", taxCode)
		}

		vatAmount := totalSales * float64(taxCode) / 100.0

		entry := domain.VATReportEntry{
			TaxCode:    taxCode,
			TaxRate:    label,
			TotalSales: totalSales,
			TotalVAT:   vatAmount,
		}

		report.Entries = append(report.Entries, entry)
		report.TotalSales += totalSales
		report.TotalVAT += vatAmount
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return report, nil
}
