// User types
export type UserRole = "Admin" | "Bookkeeper" | "Manager";

export interface User {
  user_id: number;
  first_name: string;
  last_name: string;
  name: string;
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

// Line Item types
export interface LineItem {
  line_id: number;
  voucher_id: number;
  account_no: number;
  debit_amount: number;
  credit_amount: number;
  tax_code: number; // 0, 6, 12, 25
  project_id?: number;
  cost_center_id?: number;
  account?: Account; // Populated in responses
}

export interface CreateLineItemRequest {
  voucher_id: number;
  account_no: number;
  debit_amount: number;
  credit_amount: number;
  tax_code: number;
  project_id?: number;
  cost_center_id?: number;
}

export interface UpdateLineItemRequest {
  account_no?: number;
  debit_amount?: number;
  credit_amount?: number;
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
  total_amount: number;
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
  total_amount: number;
  period: string;
  created_by: number;
  lines?: CreateLineItemRequest[];
}

export interface UpdateVoucherRequest {
  date?: string;
  description?: string;
  reference?: string;
  total_amount?: number;
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

// API Response types
export interface ApiError {
  error: string;
  message?: string;
}

export interface ValidationResponse {
  balanced: boolean;
  total_debit: number;
  total_credit: number;
  difference: number;
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
  debit_amount: number;
  credit_amount: number;
  balance: number;
}
