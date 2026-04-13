import { InvoiceSettings, UpdateInvoiceSettingsRequest } from "@/types";
import { apiClient } from "./client";

export const invoicesApi = {
  getSettings: async (): Promise<InvoiceSettings> =>
    apiClient.get("/invoices/settings"),
  updateSettings: async (
    data: UpdateInvoiceSettingsRequest
  ): Promise<InvoiceSettings> => apiClient.put("/invoices/settings", data),
};
