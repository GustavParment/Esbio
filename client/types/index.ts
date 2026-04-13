// User types
export type UserRole = "Admin" | "Bookkeeper" | "Manager";

export interface User {
  user_id: number;
  first_name: string;
  last_name: string;
  name: string;
  company_name?: string;
  org_number?: string;
  email: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  role: UserRole;
}

export interface UpdateUserRequest {
  name?: string;
  email?: string;
  password?: string;
  role?: UserRole;
  company_name?: string;
  org_number?: string;
}

// Company types
export interface Company {
  company_id: number;
  company_name: string;
  org_number?: string;
  plan: string;
  plan_status: string;
  created_by: number;
  created_at: string;
  updated_at: string;
  trial_days_left?: number;
  stripe_customer_id?: string;
  stripe_subscription_id?: string;
}

export interface CreateCompanyRequest {
  company_name: string;
  org_number?: string;
}

// Account types
export type AccountType = "P&L" | "BS";
export type StandardSide = "Debit" | "Credit";
export type TaxStandard = "25%" | "12%" | "6%" | "0%";

export interface Account {
  account_no: number;
  account_name: string;
  account_group: number; // 1-8
  tax_standard: TaxStandard;
  type: AccountType;
  standard_side: StandardSide;
}

export interface CreateAccountRequest {
  account_no: number;
  account_name: string;
  account_group: number;
  tax_standard: TaxStandard;
  type: AccountType;
  standard_side: StandardSide;
}

export interface UpdateAccountRequest {
  account_name?: string;
  account_group?: number;
  tax_standard?: TaxStandard;
  type?: AccountType;
  standard_side?: StandardSide;
}

// Money values are transported as decimal strings (e.g. "1234.56") to
// preserve exact precision — JS number cannot represent 2-decimal values
// past 2^53 and rounds subtotals (0.1+0.2 !== 0.3). Parse with parseFloat
// for display, or use a Decimal library for client-side math.
export type MoneyString = string;

// Line Item types
export interface LineItem {
  line_id: number;
  voucher_id: number;
  account_no: number;
  debit_amount: MoneyString;
  credit_amount: MoneyString;
  tax_code: number; // 0, 6, 12, 25
  project_id?: number;
  cost_center_id?: number;
  account?: Account; // Populated in responses
}

export interface CreateLineItemRequest {
  voucher_id: number;
  account_no: number;
  debit_amount: MoneyString | number;
  credit_amount: MoneyString | number;
  tax_code: number;
  project_id?: number;
  cost_center_id?: number;
}

export interface UpdateLineItemRequest {
  account_no?: number;
  debit_amount?: MoneyString | number;
  credit_amount?: MoneyString | number;
  tax_code?: number;
  project_id?: number;
  cost_center_id?: number;
}

// Voucher types
export interface Voucher {
  voucher_id: number;
  voucher_number: number;
  date: string;
  description: string;
  reference: string;
  total_amount: MoneyString;
  period: string; // YYYY-MM
  created_by: number;
  corrects_voucher_id?: number | null;     // ID för verifikat som detta rättar
  corrected_by_voucher_id?: number | null; // ID för verifikat som rättat detta
  created_at: string;
  updated_at: string;
  lines?: LineItem[]; // Populated in detail responses
}

export interface CreateVoucherRequest {
  date: string;
  description: string;
  reference: string;
  total_amount: MoneyString | number;
  period: string;
  created_by: number;
  lines?: CreateLineItemRequest[];
}

export interface UpdateVoucherRequest {
  date?: string;
  description?: string;
  reference?: string;
  total_amount?: MoneyString | number;
  period?: string;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  role: UserRole;
}

export interface RegisterResponse {
  message: string;
  user: User;
}

// Receipt scanning types
export interface ReceiptLineItem {
  account_no: number;
  account_name: string;
  debit_amount: MoneyString;
  credit_amount: MoneyString;
  tax_code: number;
}

export interface ReceiptScanResult {
  date: string;
  vendor: string;
  description: string;
  total_amount: MoneyString;
  vat_rate: number;
  vat_amount: MoneyString;
  amount_excl_vat: MoneyString;
  currency: string;
  suggested_lines: ReceiptLineItem[];
}

// API Response types
export interface ApiError {
  error: string;
  message?: string;
}

export interface ValidationResponse {
  balanced: boolean;
  total_debit: MoneyString;
  total_credit: MoneyString;
  difference: MoneyString;
}

// Pagination and filtering
export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface AccountFilters extends PaginationParams {
  group?: number;
  type?: AccountType;
}

export interface VoucherFilters extends PaginationParams {
  period?: string;
  user_id?: number;
  from_date?: string;
  to_date?: string;
}

// Ledger types
export interface LedgerEntry {
  date: string;
  voucher_id: number;
  voucher_number: number;
  description: string;
  reference: string;
  debit_amount: MoneyString;
  credit_amount: MoneyString;
  balance: MoneyString;
}

// Customer types
export interface Customer {
  customer_id: number;
  company_id: number;
  name: string;
  org_number?: string;
  vat_number?: string;
  address_line1?: string;
  address_line2?: string;
  postal_code?: string;
  city?: string;
  country?: string;
  email?: string;
  phone?: string;
  payment_terms_days: number;
  default_revenue_account?: number;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCustomerRequest {
  name: string;
  org_number?: string;
  vat_number?: string;
  address_line1?: string;
  address_line2?: string;
  postal_code?: string;
  city?: string;
  country?: string;
  email?: string;
  phone?: string;
  payment_terms_days?: number;
  default_revenue_account?: number;
  notes?: string;
}

// Invoice settings types
export interface InvoiceSettings {
  settings_id: number;
  company_id: number;
  bankgiro?: string;
  plusgiro?: string;
  swish?: string;
  iban?: string;
  bic?: string;
  f_skatt_text: string;
  default_payment_terms_days: number;
  next_invoice_number: number;
  invoice_prefix?: string;
  default_revenue_account: number;
  default_payment_account: number;
  footer_text?: string;
}

export interface UpdateInvoiceSettingsRequest {
  bankgiro?: string;
  plusgiro?: string;
  swish?: string;
  iban?: string;
  bic?: string;
  f_skatt_text?: string;
  default_payment_terms_days?: number;
  invoice_prefix?: string;
  default_revenue_account?: number;
  default_payment_account?: number;
  footer_text?: string;
}
