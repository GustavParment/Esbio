import {
  Invoice,
  CreateInvoiceRequest,
  InvoiceSettings,
  UpdateInvoiceSettingsRequest,
  MarkAsPaidRequest,
} from "@/types";
import { apiClient } from "./client";

export const invoicesApi = {
  getAll: async (params?: {
    status?: string;
    customer_id?: number;
  }): Promise<Invoice[]> => {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.customer_id)
      query.set("customer_id", String(params.customer_id));
    const qs = query.toString();
    return apiClient.get(`/invoices${qs ? "?" + qs : ""}`);
  },
  getById: async (id: number): Promise<Invoice> =>
    apiClient.get(`/invoices/${id}`),
  create: async (data: CreateInvoiceRequest): Promise<Invoice> =>
    apiClient.post("/invoices", data),
  finalize: async (id: number): Promise<Invoice> =>
    apiClient.post(`/invoices/${id}/finalize`, {}),
  markAsPaid: async (id: number, data: MarkAsPaidRequest): Promise<Invoice> =>
    apiClient.post(`/invoices/${id}/pay`, data),
  cancel: async (id: number): Promise<Invoice> =>
    apiClient.post(`/invoices/${id}/cancel`, {}),
  delete: async (id: number): Promise<void> =>
    apiClient.delete(`/invoices/${id}`),
  sendEmail: async (id: number): Promise<{ message: string; email_id: string }> =>
    apiClient.post(`/invoices/${id}/send-email`, {}),
  getSettings: async (): Promise<InvoiceSettings> =>
    apiClient.get("/invoices/settings"),
  updateSettings: async (
    data: UpdateInvoiceSettingsRequest
  ): Promise<InvoiceSettings> => apiClient.put("/invoices/settings", data),
};
