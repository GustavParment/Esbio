import { ReceiptScanResult } from "@/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "https://api.esbio.se/api/v1";

export const receiptsApi = {
  scan: async (file: File): Promise<ReceiptScanResult> => {
    const formData = new FormData();
    formData.append("receipt", file);

    const response = await fetch(`${API_BASE_URL}/receipts/scan`, {
      method: "POST",
      credentials: "include",
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Scan failed" }));
      throw new Error(error.error || "Failed to scan receipt");
    }

    return response.json();
  },
};
