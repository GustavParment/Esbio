import { apiClient } from "./client";
import { Company, CreateCompanyRequest } from "@/types";

export const companiesApi = {
  list: async (): Promise<Company[]> => {
    return apiClient.get<Company[]>("/companies");
  },

  create: async (data: CreateCompanyRequest): Promise<Company> => {
    return apiClient.post<Company>("/companies", data);
  },

  select: async (companyId: number): Promise<void> => {
    await apiClient.post("/companies/select", { company_id: companyId });
  },

  update: async (id: number, data: Partial<CreateCompanyRequest>): Promise<Company> => {
    return apiClient.put<Company>(`/companies/${id}`, data);
  },

  delete: async (id: number): Promise<void> => {
    return apiClient.delete<void>(`/companies/${id}`);
  },
};
