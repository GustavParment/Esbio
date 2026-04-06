import { Voucher, CreateVoucherRequest, UpdateVoucherRequest, ValidationResponse } from "@/types";
import { apiClient } from "./client";

export const vouchersApi = {
  getAll: async (): Promise<Voucher[]> => {
    return apiClient.get<Voucher[]>("/vouchers");
  },

  getById: async (id: number): Promise<Voucher> => {
    return apiClient.get<Voucher>(`/vouchers/${id}`);
  },

  getByPeriod: async (period: string): Promise<Voucher[]> => {
    return apiClient.get<Voucher[]>(`/vouchers/period/${period}`);
  },

  getByCompany: async (): Promise<Voucher[]> => {
    return apiClient.get<Voucher[]>("/vouchers/company");
  },

  getAllPeriods: async (): Promise<string[]> => {
    return apiClient.get<string[]>("/vouchers/periods");
  },

  validate: async (id: number): Promise<ValidationResponse> => {
    return apiClient.get<ValidationResponse>(`/vouchers/${id}/validate`);
  },

  create: async (data: CreateVoucherRequest): Promise<Voucher> => {
    return apiClient.post<Voucher>("/vouchers", data);
  },

  update: async (id: number, data: UpdateVoucherRequest): Promise<Voucher> => {
    return apiClient.put<Voucher>(`/vouchers/${id}`, data);
  },

  delete: async (id: number): Promise<void> => {
    return apiClient.delete<void>(`/vouchers/${id}`);
  },

  createCorrection: async (id: number, userId: number): Promise<Voucher> => {
    return apiClient.post<Voucher>(`/vouchers/${id}/correct`, { user_id: userId });
  },

  getPdf: async (id: number): Promise<Blob> => {
    const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || "https://api.esbio.se/api/v1"}/vouchers/${id}/pdf`, {
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error("Failed to download PDF");
    }
    return response.blob();
  },
};
