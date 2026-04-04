import { apiClient } from "./client";

export interface IncomeStatementEntry {
  account_no: number;
  account_name: string;
  balance: number;
}

export interface IncomeStatement {
  period: {
    from_date: string;
    to_date: string;
  };
  income: IncomeStatementEntry[];
  expenses: IncomeStatementEntry[];
  total_income: number;
  total_expenses: number;
  net_result: number;
}

export interface BalanceSheetEntry {
  account_no: number;
  account_name: string;
  balance: number;
}

export interface BalanceSheet {
  as_of_date: string;
  assets: BalanceSheetEntry[];
  equity_liabilities: BalanceSheetEntry[];
  total_assets: number;
  total_equity_liabilities: number;
  net_result: number;
}

export interface VATReportEntry {
  tax_code: number;
  tax_rate: string;
  total_sales: number;
  total_vat: number;
}

export interface VATReport {
  period: {
    from_date: string;
    to_date: string;
  };
  entries: VATReportEntry[];
  total_sales: number;
  total_vat: number;
}

export const reportsApi = {
  getIncomeStatement: async (fromDate: string, toDate: string): Promise<IncomeStatement> => {
    return apiClient.get<IncomeStatement>(`/reports/income-statement?from_date=${fromDate}&to_date=${toDate}`);
  },

  getBalanceSheet: async (asOfDate: string): Promise<BalanceSheet> => {
    return apiClient.get<BalanceSheet>(`/reports/balance-sheet?as_of_date=${asOfDate}`);
  },

  getVATReport: async (fromDate: string, toDate: string): Promise<VATReport> => {
    return apiClient.get<VATReport>(`/reports/vat?from_date=${fromDate}&to_date=${toDate}`);
  },
};
