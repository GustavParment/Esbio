import { Customer, CreateCustomerRequest } from "@/types";
import { apiClient } from "./client";

export const customersApi = {
  getAll: async (): Promise<Customer[]> => apiClient.get("/customers"),
  getById: async (id: number): Promise<Customer> =>
    apiClient.get(`/customers/${id}`),
  search: async (query: string): Promise<Customer[]> =>
    apiClient.get(`/customers/search?q=${encodeURIComponent(query)}`),
  create: async (data: CreateCustomerRequest): Promise<Customer> =>
    apiClient.post("/customers", data),
  update: async (
    id: number,
    data: Partial<CreateCustomerRequest>
  ): Promise<Customer> => apiClient.put(`/customers/${id}`, data),
  delete: async (id: number): Promise<void> =>
    apiClient.delete(`/customers/${id}`),
};
