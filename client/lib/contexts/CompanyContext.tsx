"use client";

import { createContext, useContext, useEffect, useState, useCallback } from "react";
import { Company } from "@/types";
import { companiesApi } from "@/lib/api/companies";
import { useAuth } from "./AuthContext";

interface CompanyContextType {
  companies: Company[];
  selectedCompany: Company | null;
  loading: boolean;
  selectCompany: (companyId: number) => Promise<void>;
  refreshCompanies: () => Promise<void>;
}

const CompanyContext = createContext<CompanyContextType>({
  companies: [],
  selectedCompany: null,
  loading: true,
  selectCompany: async () => {},
  refreshCompanies: async () => {},
});

export function CompanyProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const [companies, setCompanies] = useState<Company[]>([]);
  const [selectedCompany, setSelectedCompany] = useState<Company | null>(() => {
    // Sync restore from localStorage to prevent ProtectedRoute redirect before API loads
    if (typeof window !== "undefined") {
      const savedId = localStorage.getItem("selectedCompanyId");
      if (savedId) {
        return { company_id: parseInt(savedId) } as Company;
      }
    }
    return null;
  });
  const [loading, setLoading] = useState(true);

  const refreshCompanies = useCallback(async () => {
    if (!user) {
      setCompanies([]);
      // Don't clear selectedCompany or localStorage here — auth might still be loading.
      // These get cleared on explicit logout (user goes from set → null).
      setLoading(false);
      return;
    }

    setLoading(true);
    try {
      const data = await companiesApi.list();
      const list = Array.isArray(data) ? data : [];
      setCompanies(list);

      // If we have a selected company, refresh it from the list
      if (selectedCompany) {
        const updated = list.find(c => c.company_id === selectedCompany.company_id);
        if (updated) setSelectedCompany(updated);
      } else {
        // Restore selected company from localStorage (survives page reloads + external redirects)
        const savedId = localStorage.getItem("selectedCompanyId");
        if (savedId) {
          const id = parseInt(savedId);
          const company = list.find(c => c.company_id === id);
          if (company) setSelectedCompany(company);
        }
      }
    } catch {
      setCompanies([]);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    refreshCompanies();
  }, [user]);

  const selectCompany = async (companyId: number) => {
    try {
      await companiesApi.select(companyId);
      const company = companies.find(c => c.company_id === companyId);
      if (company) {
        setSelectedCompany(company);
        localStorage.setItem("selectedCompanyId", String(companyId));
      }
    } catch (err) {
      console.error("Failed to select company:", err);
    }
  };

  return (
    <CompanyContext.Provider value={{ companies, selectedCompany, loading, selectCompany, refreshCompanies }}>
      {children}
    </CompanyContext.Provider>
  );
}

export const useCompany = () => useContext(CompanyContext);
