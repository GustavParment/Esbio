"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/contexts/AuthContext";
import { useCompany } from "@/lib/contexts/CompanyContext";
import Link from "next/link";

export default function CompanySelectorPage() {
  const { user, loading: authLoading, logout } = useAuth();
  const { companies, loading: compLoading, selectCompany, refreshCompanies } = useCompany();
  const router = useRouter();

  useEffect(() => {
    if (!authLoading && !user) {
      router.push("/auth/login");
    }
  }, [user, authLoading, router]);

  useEffect(() => {
    if (user) refreshCompanies();
  }, [user]);

  const handleSelect = async (companyId: number) => {
    await selectCompany(companyId);
    router.push("/dashboard");
  };

  if (authLoading || compLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
      <div className="max-w-3xl mx-auto px-4 py-12">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Mina Företag</h1>
          </div>
          <p className="text-gray-600 dark:text-gray-400 mt-1">Välj ett företag att arbeta med, eller skapa ett nytt</p>
          <Link
            href="/companies/new"
            className="mt-4 block w-full text-center bg-blue-600 hover:bg-blue-700 text-white font-medium py-3 px-4 rounded-lg transition-colors"
          >
            + Nytt företag
          </Link>
        </div>

        {/* Company list */}
        {companies.length === 0 ? (
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
            <div className="text-5xl mb-4">🏢</div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Inget företag skapat</h2>
            <p className="text-gray-500 mb-6">Du behöver skapa ett företag innan du kan börja bokföra</p>
            <Link
              href="/companies/new"
              className="px-6 py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 transition-colors font-medium"
            >
              Skapa ditt första företag
            </Link>
          </div>
        ) : (
          <div className="space-y-3">
            {companies.map((company) => (
              <button
                key={company.company_id}
                onClick={() => handleSelect(company.company_id)}
                className="w-full bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5 hover:border-blue-300 dark:hover:border-blue-600 hover:shadow-md transition-all text-left flex items-center justify-between group"
              >
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-blue-600 transition-colors">
                    {company.company_name}
                  </h3>
                  <div className="flex gap-4 mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {company.org_number && <span>Org.nr: {company.org_number}</span>}
                    <span className="capitalize">
                      {company.plan === "free" ? "Gratis" : company.plan === "starter" ? "Starter" : company.plan === "growth" ? "Tillväxt" : company.plan}
                    </span>
                  </div>
                </div>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-gray-400 group-hover:text-blue-600 transition-colors">
                  <polyline points="9 18 15 12 9 6"/>
                </svg>
              </button>
            ))}
          </div>
        )}

        {/* Footer */}
        <div className="mt-8 text-center">
          <button
            onClick={() => logout()}
            className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
          >
            Logga ut
          </button>
        </div>
      </div>
    </div>
  );
}
