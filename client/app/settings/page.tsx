"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { useAuth } from "@/lib/contexts/AuthContext";
import { useCompany } from "@/lib/contexts/CompanyContext";
import { companiesApi } from "@/lib/api/companies";
import { authApi } from "@/lib/api/auth";
import { agentApi, AgentUsageSummary } from "@/lib/api/agent";
import { useRouter } from "next/navigation";

export default function SettingsPage() {
  const { user, logout } = useAuth();
  const { selectedCompany, refreshCompanies } = useCompany();
  const router = useRouter();
  const [companyName, setCompanyName] = useState("");
  const [orgNumber, setOrgNumber] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [usage, setUsage] = useState<AgentUsageSummary | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState("");

  useEffect(() => {
    if (selectedCompany) {
      setCompanyName(selectedCompany.company_name || "");
      setOrgNumber(selectedCompany.org_number || "");
    }
  }, [selectedCompany]);

  useEffect(() => {
    const fetchUsage = async () => {
      try {
        const data = await agentApi.getUsage();
        setUsage(data);
      } catch {
        // Ignore
      }
    };
    fetchUsage();
  }, []);

  const handleSave = async () => {
    if (!selectedCompany) return;
    setLoading(true);
    setError("");
    setSuccess("");
    try {
      await companiesApi.update(selectedCompany.company_id, {
        company_name: companyName,
        org_number: orgNumber,
      });
      await refreshCompanies();
      setSuccess("Inställningar sparade!");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte spara");
    } finally {
      setLoading(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Inställningar</h1>
        <p className="text-gray-600 mt-2">Hantera företaget och se AI-användning</p>
      </div>

      {/* Company Info */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Företagsinformation</h2>
        <p className="text-sm text-gray-500 mb-4">
          Denna information används i SIE4-export och rapporter
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div>
            <label htmlFor="company_name" className="block text-sm font-medium text-gray-700 mb-2">
              Företagsnamn
            </label>
            <input
              id="company_name"
              type="text"
              value={companyName}
              onChange={(e) => setCompanyName(e.target.value)}
              placeholder="Mitt Företag AB"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
            />
          </div>
          <div>
            <label htmlFor="org_number" className="block text-sm font-medium text-gray-700 mb-2">
              Organisationsnummer
            </label>
            <input
              id="org_number"
              type="text"
              value={orgNumber}
              onChange={(e) => setOrgNumber(e.target.value)}
              placeholder="559123-4567"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
            />
          </div>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700 text-sm">{error}</p>
          </div>
        )}

        {success && (
          <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg">
            <p className="text-green-700 text-sm">{success}</p>
          </div>
        )}

        <button
          onClick={handleSave}
          disabled={loading}
          className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50"
        >
          {loading ? "Sparar..." : "Spara"}
        </button>
      </div>

      {/* AI Usage */}
      {usage && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Ester AI — Användning</h2>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-sm text-gray-500">Denna månad</p>
              <p className="text-xl font-bold text-gray-900">{usage.month_request_count}</p>
              <p className="text-xs text-gray-500">förfrågningar</p>
            </div>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-sm text-gray-500">Tokens denna månad</p>
              <p className="text-xl font-bold text-gray-900">{usage.month_total_tokens.toLocaleString("sv-SE")}</p>
              <p className="text-xs text-gray-500">tokens</p>
            </div>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-sm text-gray-500">Kostnad denna månad</p>
              <p className="text-xl font-bold text-gray-900">
                {usage.month_estimated_cost_sek.toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr
              </p>
              <p className="text-xs text-gray-500">uppskattad</p>
            </div>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-sm text-gray-500">Totalt alla tider</p>
              <p className="text-xl font-bold text-gray-900">
                {usage.estimated_cost_sek.toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr
              </p>
              <p className="text-xs text-gray-500">{usage.request_count} förfrågningar</p>
            </div>
          </div>
        </div>
      )}

      {/* User Info */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Kontoinformation</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-gray-500">Namn</p>
            <p className="font-medium text-gray-900">{user?.first_name} {user?.last_name}</p>
          </div>
          <div>
            <p className="text-gray-500">E-post</p>
            <p className="font-medium text-gray-900">{user?.email}</p>
          </div>
          <div>
            <p className="text-gray-500">Roll</p>
            <p className="font-medium text-gray-900">{user?.role}</p>
          </div>
        </div>
      </div>

      {/* Delete Account */}
      <div className="bg-white rounded-xl shadow-sm border-2 border-red-200 p-6">
        <h2 className="text-lg font-semibold text-red-600 mb-2">Radera konto</h2>
        <p className="text-sm text-gray-600 mb-4">
          Detta raderar ditt konto och all data permanent — alla företag, verifikat, transaktioner och rapporter.
          Denna åtgärd kan inte ångras.
        </p>

        {deleteError && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700 text-sm">{deleteError}</p>
          </div>
        )}

        <div className="flex items-end gap-4">
          <div className="flex-1 max-w-xs">
            <label htmlFor="delete_confirm" className="block text-sm font-medium text-gray-700 mb-2">
              Skriv <span className="font-mono font-bold">RADERA</span> för att bekräfta
            </label>
            <input
              id="delete_confirm"
              type="text"
              value={deleteConfirm}
              onChange={(e) => setDeleteConfirm(e.target.value)}
              placeholder="RADERA"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
            />
          </div>
          <button
            onClick={async () => {
              if (deleteConfirm !== "RADERA") {
                setDeleteError("Du måste skriva RADERA för att bekräfta.");
                return;
              }
              setDeleting(true);
              setDeleteError("");
              try {
                await authApi.deleteAccount();
                await logout();
                router.push("/auth/login");
              } catch (err) {
                setDeleteError(err instanceof Error ? err.message : "Kunde inte radera kontot");
              } finally {
                setDeleting(false);
              }
            }}
            disabled={deleting || deleteConfirm !== "RADERA"}
            className="px-6 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {deleting ? "Raderar..." : "Radera mitt konto"}
          </button>
        </div>
      </div>
    </DashboardLayout>
  );
}
