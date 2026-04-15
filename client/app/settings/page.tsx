"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { useAuth } from "@/lib/contexts/AuthContext";
import { useCompany } from "@/lib/contexts/CompanyContext";
import { companiesApi } from "@/lib/api/companies";
import { authApi } from "@/lib/api/auth";
import { stripeApi } from "@/lib/api/stripe";
import { useRouter } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import Link from "next/link";
import SIEImportSection from "@/components/settings/SIEImportSection";
import { formatOrgNumber, validateOrgNumber } from "@/lib/utils/orgNumber";

export default function SettingsPage() {
  const { user, logout } = useAuth();
  const { selectedCompany, refreshCompanies } = useCompany();
  const router = useRouter();
  const [companyName, setCompanyName] = useState("");
  const [orgNumber, setOrgNumber] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [deleteConfirm, setDeleteConfirm] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState("");
  const [supportSubject, setSupportSubject] = useState("");
  const [supportMessage, setSupportMessage] = useState("");
  const [supportLoading, setSupportLoading] = useState(false);
  const [supportStatus, setSupportStatus] = useState<"" | "success" | "error">("");
  const [orgError, setOrgError] = useState<string | null>(null);

  useEffect(() => {
    if (selectedCompany) {
      setCompanyName(selectedCompany.company_name || "");
      setOrgNumber(selectedCompany.org_number || "");
    }
  }, [selectedCompany]);

  const handleSave = async () => {
    if (!selectedCompany) return;
    const orgErr = validateOrgNumber(orgNumber);
    if (orgErr) {
      setOrgError(orgErr);
      return;
    }
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
              inputMode="numeric"
              value={orgNumber}
              onChange={(e) => {
                setOrgNumber(formatOrgNumber(e.target.value));
                setOrgError(null);
              }}
              onBlur={() => setOrgError(validateOrgNumber(orgNumber))}
              placeholder="559123-4567"
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:border-transparent text-gray-900 placeholder:text-gray-400 ${
                orgError ? "border-red-400 focus:ring-red-500" : "border-gray-300 focus:ring-blue-500"
              }`}
            />
            {orgError && (
              <p className="mt-1 text-xs text-red-600">{orgError}</p>
            )}
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
          className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm disabled:opacity-50"
        >
          {loading ? "Sparar..." : "Spara"}
        </button>
      </div>

      {/* SIE Import */}
      <SIEImportSection />

      {/* Subscription */}
      {selectedCompany && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Prenumeration</h2>
          <div>
            <div className="flex flex-wrap items-center gap-2 mb-1">
              <span className="text-2xl font-bold text-gray-900">
                {selectedCompany.plan === "free" && "Gratis (provperiod)"}
                {selectedCompany.plan === "starter" && "Starter"}
                {selectedCompany.plan === "growth" && "Tillväxt"}
              </span>
              {selectedCompany.plan !== "free" && (
                <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  selectedCompany.plan_status === "active"
                    ? "bg-green-100 text-green-700"
                    : selectedCompany.plan_status === "past_due"
                    ? "bg-red-100 text-red-700"
                    : "bg-gray-100 text-gray-700"
                }`}>
                  {selectedCompany.plan_status === "active" && "Aktiv"}
                  {selectedCompany.plan_status === "past_due" && "Betalning förfallen"}
                  {selectedCompany.plan_status === "canceled" && "Avslutad"}
                  {selectedCompany.plan_status === "trialing" && "Provperiod"}
                </span>
              )}
            </div>
            <p className="text-sm text-gray-500 mb-4">
              {selectedCompany.plan === "free" && "Uppgradera för att låsa upp alla funktioner"}
              {selectedCompany.plan === "starter" && "199 kr/mån"}
              {selectedCompany.plan === "growth" && "399 kr/mån"}
            </p>
            <div className="flex flex-wrap gap-2">
              {selectedCompany.plan === "free" && (
                <Link
                  href="/upgrade"
                  className="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm"
                >
                  Uppgradera
                </Link>
              )}
              {selectedCompany.plan === "starter" && (
                <Link
                  href="/upgrade"
                  className="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm"
                >
                  Uppgradera till Tillväxt
                </Link>
              )}
              {selectedCompany.plan !== "free" && (
                <button
                  onClick={async () => {
                    try {
                      const { url } = await stripeApi.createPortalSession();
                      window.location.href = url;
                    } catch (err) {
                      setError(err instanceof Error ? err.message : "Kunde inte öppna kundportalen");
                    }
                  }}
                  className="px-5 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium text-sm"
                >
                  Hantera prenumeration
                </button>
              )}
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

      {/* Support */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-2">Kontakta support</h2>
        <p className="text-sm text-gray-500 mb-4">Har du frågor eller behöver hjälp? Skicka ett meddelande så återkommer vi.</p>

        {supportStatus === "success" ? (
          <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
            <p className="text-green-700 font-medium">Tack! Vi har tagit emot ditt meddelande och återkommer så snart vi kan.</p>
            <button
              onClick={() => { setSupportStatus(""); setSupportSubject(""); setSupportMessage(""); }}
              className="mt-2 text-sm text-green-600 hover:text-green-700 font-medium"
            >
              Skicka ett till
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <label htmlFor="support_subject" className="block text-sm font-medium text-gray-700 mb-1">
                Ämne
              </label>
              <input
                id="support_subject"
                type="text"
                value={supportSubject}
                onChange={(e) => setSupportSubject(e.target.value)}
                placeholder="Vad gäller det?"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
              />
            </div>
            <div>
              <label htmlFor="support_message" className="block text-sm font-medium text-gray-700 mb-1">
                Meddelande
              </label>
              <textarea
                id="support_message"
                rows={4}
                value={supportMessage}
                onChange={(e) => setSupportMessage(e.target.value)}
                placeholder="Beskriv ditt ärende..."
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400 resize-none"
              />
            </div>

            {supportStatus === "error" && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-red-700 text-sm">Kunde inte skicka meddelandet. Försök igen eller maila info@esbio.se direkt.</p>
              </div>
            )}

            <button
              onClick={async () => {
                if (!supportSubject.trim() || !supportMessage.trim()) return;
                setSupportLoading(true);
                setSupportStatus("");
                try {
                  await apiClient.post("/support", { subject: supportSubject, message: supportMessage });
                  setSupportStatus("success");
                } catch {
                  setSupportStatus("error");
                } finally {
                  setSupportLoading(false);
                }
              }}
              disabled={supportLoading || !supportSubject.trim() || !supportMessage.trim()}
              className="px-6 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {supportLoading ? "Skickar..." : "Skicka meddelande"}
            </button>
          </div>
        )}
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
