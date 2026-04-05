"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { companiesApi } from "@/lib/api/companies";
import { useCompany } from "@/lib/contexts/CompanyContext";
import Link from "next/link";

export default function CreateCompanyPage() {
  const [companyName, setCompanyName] = useState("");
  const [orgNumber, setOrgNumber] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const { refreshCompanies, selectCompany } = useCompany();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!companyName.trim()) {
      setError("Företagsnamn krävs");
      return;
    }

    setLoading(true);
    setError("");
    try {
      const company = await companiesApi.create({
        company_name: companyName.trim(),
        org_number: orgNumber.trim() || undefined,
      });
      await refreshCompanies();
      await selectCompany(company.company_id);
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte skapa företag");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950 px-4">
      <div className="max-w-md w-full">
        <div className="bg-white rounded-2xl shadow-xl p-8">
          <Link href="/companies" className="text-blue-600 hover:text-blue-700 font-medium text-sm mb-6 inline-block">
            &larr; Tillbaka
          </Link>

          <div className="text-center mb-8">
            <img src="/eskio-logo.png" alt="Eskio" className="h-12 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-gray-900">Skapa nytt företag</h1>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label htmlFor="company_name" className="block text-sm font-medium text-gray-700 mb-2">
                Företagsnamn
              </label>
              <input
                id="company_name"
                type="text"
                required
                value={companyName}
                onChange={(e) => setCompanyName(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
                placeholder="Mitt Företag AB"
                style={{ fontSize: "16px" }}
              />
            </div>

            <div>
              <label htmlFor="org_number" className="block text-sm font-medium text-gray-700 mb-2">
                Organisationsnummer (valfritt)
              </label>
              <input
                id="org_number"
                type="text"
                value={orgNumber}
                onChange={(e) => setOrgNumber(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
                placeholder="559123-4567"
                style={{ fontSize: "16px" }}
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-3 rounded-lg transition-colors disabled:opacity-50"
            >
              {loading ? "Skapar..." : "Skapa företag"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
