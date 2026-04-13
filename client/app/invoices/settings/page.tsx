"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { invoicesApi } from "@/lib/api/invoices";
import { InvoiceSettings } from "@/types";
import Link from "next/link";

export default function InvoiceSettingsPage() {
  const [settings, setSettings] = useState<InvoiceSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const [form, setForm] = useState({
    bankgiro: "", plusgiro: "", swish: "", iban: "", bic: "",
    f_skatt_text: "Godkänd för F-skatt",
    default_payment_terms_days: 30,
    invoice_prefix: "",
    default_revenue_account: 3010,
    default_payment_account: 1930,
    footer_text: "",
  });

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const data = await invoicesApi.getSettings();
        setSettings(data);
        setForm({
          bankgiro: data.bankgiro || "", plusgiro: data.plusgiro || "",
          swish: data.swish || "", iban: data.iban || "", bic: data.bic || "",
          f_skatt_text: data.f_skatt_text || "Godkänd för F-skatt",
          default_payment_terms_days: data.default_payment_terms_days,
          invoice_prefix: data.invoice_prefix || "",
          default_revenue_account: data.default_revenue_account,
          default_payment_account: data.default_payment_account,
          footer_text: data.footer_text || "",
        });
      } catch (err) {
        setError("Kunde inte hämta inställningar");
      } finally {
        setLoading(false);
      }
    };
    fetchSettings();
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const updated = await invoicesApi.updateSettings({
        ...form,
        default_payment_terms_days: Number(form.default_payment_terms_days) || 30,
        default_revenue_account: Number(form.default_revenue_account) || 3010,
        default_payment_account: Number(form.default_payment_account) || 1930,
      });
      setSettings(updated);
      setSuccess("Inställningar sparade");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte spara");
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="mb-8">
        <Link href="/invoices" className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block">
          &larr; Tillbaka till fakturor
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">Fakturainställningar</h1>
        <p className="text-gray-600 mt-2">Betalningsuppgifter och standardvärden för nya fakturor</p>
      </div>

      {error && <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">{error}</div>}
      {success && <div className="mb-6 p-4 bg-green-50 border border-green-200 text-green-700 rounded-lg">{success}</div>}

      <form onSubmit={handleSubmit}>
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Betalningsuppgifter</h2>
          <p className="text-sm text-gray-500 mb-4">Dessa visas på fakturan som betalningsinformation</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Bankgiro</label>
              <input name="bankgiro" value={form.bankgiro} onChange={handleChange} placeholder="1234-5678"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Plusgiro</label>
              <input name="plusgiro" value={form.plusgiro} onChange={handleChange} placeholder="12 34 56-7"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Swish</label>
              <input name="swish" value={form.swish} onChange={handleChange} placeholder="1234567890"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">IBAN</label>
              <input name="iban" value={form.iban} onChange={handleChange} placeholder="SE1234567890123456789012"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">BIC</label>
              <input name="bic" value={form.bic} onChange={handleChange} placeholder="NDEASESS"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Standardvärden</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Betalningsvillkor (dagar)</label>
              <input name="default_payment_terms_days" type="number" value={form.default_payment_terms_days} onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Fakturaprefix</label>
              <input name="invoice_prefix" value={form.invoice_prefix} onChange={handleChange} placeholder="t.ex. F-"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Standardintäktskonto</label>
              <input name="default_revenue_account" type="number" value={form.default_revenue_account} onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Standardbetalningskonto</label>
              <input name="default_payment_account" type="number" value={form.default_payment_account} onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Övrigt</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">F-skattetext</label>
              <input name="f_skatt_text" value={form.f_skatt_text} onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Sidfot (valfri text längst ner på fakturan)</label>
              <textarea name="footer_text" value={form.footer_text} onChange={handleChange} rows={2}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
          </div>
        </div>

        {settings && (
          <div className="bg-gray-50 rounded-lg p-4 mb-6 text-sm text-gray-600">
            Nästa fakturanummer: <span className="font-semibold text-gray-900">{settings.next_invoice_number}</span>
          </div>
        )}

        <button type="submit" disabled={saving}
          className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50">
          {saving ? "Sparar..." : "Spara inställningar"}
        </button>
      </form>
    </DashboardLayout>
  );
}
