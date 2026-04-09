import { LoginRequest, LoginResponse, RegisterRequest, RegisterResponse, User } from "@/types";
import { apiClient } from "./client";

export const authApi = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    // Token will be set as httpOnly cookie by the server
    return apiClient.post<LoginResponse>("/auth/login", credentials);
  },

  register: async (data: RegisterRequest): Promise<RegisterResponse> => {
    return apiClient.post<RegisterResponse>("/auth/register", data);
  },

  logout: async (): Promise<void> => {
    // Call server logout endpoint to clear httpOnly cookie
    await apiClient.post<void>("/auth/logout", {});
  },

  getCurrentUser: async (): Promise<User> => {
    return apiClient.get<User>("/auth/me");
  },

  refreshToken: async (): Promise<{ token: string }> => {
    // Token will be refreshed as httpOnly cookie by the server
    return apiClient.post<{ token: string }>("/auth/refresh", {});
  },

  deleteAccount: async (): Promise<void> => {
    return apiClient.request<void>("/auth/account", {
      method: "DELETE",
      body: JSON.stringify({ confirm: "DELETE" }),
    });
  },
};
