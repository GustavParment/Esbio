import { apiClient } from "./client";
import type { MoneyString } from "@/types";

export interface IncomeStatementEntry {
  account_no: number;
  account_name: string;
  balance: MoneyString;
}

export interface IncomeStatement {
  period: {
    from_date: string;
    to_date: string;
  };
  income: IncomeStatementEntry[];
  expenses: IncomeStatementEntry[];
  total_income: MoneyString;
  total_expenses: MoneyString;
  net_result: MoneyString;
}

export interface BalanceSheetEntry {
  account_no: number;
  account_name: string;
  balance: MoneyString;
}

export interface BalanceSheet {
  as_of_date: string;
  assets: BalanceSheetEntry[];
  equity_liabilities: BalanceSheetEntry[];
  total_assets: MoneyString;
  total_equity_liabilities: MoneyString;
  net_result: MoneyString;
}

export interface VATReportEntry {
  tax_code: number;
  tax_rate: string;
  total_sales: MoneyString;
  total_vat: MoneyString;
}

export interface VATInputEntry {
  account_no: number;
  account_name: string;
  amount: MoneyString;
}

export interface VATReport {
  period: {
    from_date: string;
    to_date: string;
  };
  entries: VATReportEntry[];
  total_sales: MoneyString;
  total_vat: MoneyString;
  input_entries: VATInputEntry[];
  total_input_vat: MoneyString;
  net_vat: MoneyString;
}

export interface SIEImportResult {
  accounts_created: number;
  accounts_skipped: number;
  vouchers_created: number;
  vouchers_skipped: number;
}

export interface SIEImportSummary {
  company_name: string;
  org_number: string;
  chart_type: string;
  fiscal_from: string;
  fiscal_to: string;
  encoding: string;
  accounts_in_file: number;
  accounts_to_add: number;
  vouchers_in_file: number;
  unbalanced_count: number;
  warnings: string[];
  dry_run: boolean;
  imported?: SIEImportResult;
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

  uploadSIE: async (file: File, dryRun: boolean): Promise<SIEImportSummary> => {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || "https://api.esbio.se/api/v1";
    const path = dryRun ? "/reports/sie/import/preview" : "/reports/sie/import";
    const form = new FormData();
    form.append("file", file);
    const response = await fetch(`${baseUrl}${path}`, {
      method: "POST",
      credentials: "include",
      body: form,
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: "Import failed" }));
      throw new Error(err.error || "Failed to import SIE");
    }
    return response.json();
  },

  downloadVATDeclaration: async (fromDate: string, toDate: string): Promise<void> => {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || "https://api.esbio.se/api/v1";
    const response = await fetch(`${baseUrl}/reports/vat/pdf?from_date=${fromDate}&to_date=${toDate}`, {
      credentials: "include",
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: "Download failed" }));
      throw new Error(err.error || "Failed to download momsdeklaration");
    }
    const blob = await response.blob();
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = `momsdeklaration_${fromDate}_${toDate}.pdf`;
    a.click();
    URL.revokeObjectURL(a.href);
  },

  downloadSIE: async (fromDate: string, toDate: string): Promise<void> => {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || "https://api.esbio.se/api/v1";
    const response = await fetch(`${baseUrl}/reports/sie?from_date=${fromDate}&to_date=${toDate}`, {
      credentials: "include",
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: "Export failed" }));
      throw new Error(err.error || "Failed to download SIE file");
    }
    const blob = await response.blob();
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = `bokforing_${fromDate.replace(/-/g, "")}_${toDate.replace(/-/g, "")}.se`;
    a.click();
    URL.revokeObjectURL(a.href);
  },
};
