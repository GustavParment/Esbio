import { BankConnection, BankAccount, BankTransaction, ImportTransactionRequest } from "@/types";
import { apiClient } from "./client";

export const bankApi = {
  connect: async (): Promise<{ url: string }> => {
    return apiClient.post<{ url: string }>("/bank/connect", {});
  },

  handleCallback: async (state: string, code: string): Promise<{ success: boolean; company_id: number }> => {
    return apiClient.post<{ success: boolean; company_id: number }>("/bank/callback", { state, code });
  },

  getConnections: async (): Promise<BankConnection[]> => {
    return apiClient.get<BankConnection[]>("/bank/connections");
  },

  getAccounts: async (): Promise<BankAccount[]> => {
    return apiClient.get<BankAccount[]>("/bank/accounts");
  },

  getTransactions: async (params?: {
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<BankTransaction[]> => {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.limit) query.set("limit", String(params.limit));
    if (params?.offset) query.set("offset", String(params.offset));
    const qs = query.toString();
    return apiClient.get<BankTransaction[]>(`/bank/transactions${qs ? "?" + qs : ""}`);
  },

  syncTransactions: async (connectionId: number): Promise<{ new_count: number }> => {
    return apiClient.post<{ new_count: number }>(`/bank/sync/${connectionId}`, {});
  },

  mapAccount: async (bankAccountId: number, accountNo: number): Promise<{ success: boolean }> => {
    return apiClient.put<{ success: boolean }>(`/bank/accounts/${bankAccountId}/map`, {
      account_no: accountNo,
    });
  },

  importTransaction: async (
    txnId: number,
    data: ImportTransactionRequest
  ): Promise<{ voucher_id: number }> => {
    return apiClient.post<{ voucher_id: number }>(`/bank/transactions/${txnId}/import`, data);
  },

  skipTransaction: async (txnId: number): Promise<{ success: boolean }> => {
    return apiClient.post<{ success: boolean }>(`/bank/transactions/${txnId}/skip`, {});
  },

  disconnect: async (connectionId: number): Promise<{ success: boolean }> => {
    return apiClient.delete<{ success: boolean }>(`/bank/connections/${connectionId}`);
  },
};
