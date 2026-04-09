import { apiClient } from "./client";

export const stripeApi = {
  createCheckoutSession: async (priceId: string): Promise<{ url: string }> => {
    return apiClient.post<{ url: string }>("/stripe/checkout", { price_id: priceId });
  },

  createPortalSession: async (): Promise<{ url: string }> => {
    return apiClient.post<{ url: string }>("/stripe/portal", {});
  },
};
