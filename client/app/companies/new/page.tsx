"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { companiesApi } from "@/lib/api/companies";
import { useCompany } from "@/lib/contexts/CompanyContext";
import Link from "next/link";
import { formatOrgNumber, validateOrgNumber } from "@/lib/utils/orgNumber";

export default function CreateCompanyPage() {
  const [companyName, setCompanyName] = useState("");
  const [orgNumber, setOrgNumber] = useState("");
  const [orgError, setOrgError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const { refreshCompanies } = useCompany();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!companyName.trim()) {
      setError("Företagsnamn krävs");
      return;
    }
    const orgErr = validateOrgNumber(orgNumber);
    if (orgErr) {
      setOrgError(orgErr);
      return;
    }

    setLoading(true);
    setError("");
    try {
      await companiesApi.create({
        company_name: companyName.trim(),
        org_number: orgNumber.trim(),
      });
      await refreshCompanies();
      router.push("/companies");
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
            <img src="/esbio-logo.png" alt="Esbio" className="h-12 mx-auto mb-4" />
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
                Organisationsnummer
              </label>
              <input
                id="org_number"
                type="text"
                required
                inputMode="numeric"
                value={orgNumber}
                onChange={(e) => {
                  setOrgNumber(formatOrgNumber(e.target.value));
                  setOrgError(null);
                }}
                onBlur={() => setOrgError(validateOrgNumber(orgNumber))}
                className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:border-transparent text-gray-900 placeholder:text-gray-400 ${
                  orgError ? "border-red-400 focus:ring-red-500" : "border-gray-300 focus:ring-blue-500"
                }`}
                placeholder="559123-4567"
                style={{ fontSize: "16px" }}
              />
              {orgError && (
                <p className="mt-1 text-xs text-red-600">{orgError}</p>
              )}
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
