"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { invoicesApi } from "@/lib/api/invoices";
import { customersApi } from "@/lib/api/customers";
import { Invoice, Customer, InvoiceStatus } from "@/types";
import { formatSEK } from "@/lib/money";
import { useAuth } from "@/lib/contexts/AuthContext";
import Link from "next/link";

const STATUS_LABELS: Record<InvoiceStatus, string> = {
  draft: "Utkast",
  sent: "Skickad",
  paid: "Betald",
  overdue: "Förfallen",
  cancelled: "Makulerad",
};

const STATUS_COLORS: Record<InvoiceStatus, string> = {
  draft: "bg-gray-100 text-gray-700",
  sent: "bg-blue-100 text-blue-700",
  paid: "bg-green-100 text-green-700",
  overdue: "bg-red-100 text-red-700",
  cancelled: "bg-gray-100 text-gray-400 line-through",
};

const TABS: { label: string; value: string }[] = [
  { label: "Alla", value: "" },
  { label: "Utkast", value: "draft" },
  { label: "Skickade", value: "sent" },
  { label: "Förfallna", value: "overdue" },
  { label: "Betalda", value: "paid" },
  { label: "Makulerade", value: "cancelled" },
];

export default function InvoicesPage() {
  const { user } = useAuth();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [customers, setCustomers] = useState<Record<number, Customer>>({});
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("");

  useEffect(() => {
    if (user) fetchData();
  }, [user, activeTab]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [invData, custData] = await Promise.all([
        activeTab
          ? invoicesApi.getAll({ status: activeTab })
          : invoicesApi.getAll(),
        customersApi.getAll(),
      ]);
      setInvoices(Array.isArray(invData) ? invData : []);
      const custMap: Record<number, Customer> = {};
      for (const c of custData || []) {
        custMap[c.customer_id] = c;
      }
      setCustomers(custMap);
    } catch (error) {
      console.error("Failed to fetch invoices:", error);
    } finally {
      setLoading(false);
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
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Fakturor</h1>
            <p className="text-gray-600 mt-2">Hantera kundfakturor</p>
          </div>
          <div className="flex gap-3">
            <Link
              href="/invoices/settings"
              className="px-4 py-2.5 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors font-medium text-sm"
            >
              Inställningar
            </Link>
            <Link
              href="/invoices/new"
              className="px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm"
            >
              + Ny faktura
            </Link>
          </div>
        </div>
      </div>

      {/* Status tabs */}
      <div className="flex gap-1 mb-6 bg-gray-100 rounded-lg p-1 overflow-x-auto">
        {TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setActiveTab(tab.value)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors whitespace-nowrap ${
              activeTab === tab.value
                ? "bg-white text-gray-900 shadow-sm"
                : "text-gray-600 hover:text-gray-900"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Invoice table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {invoices.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 mb-4">
              {activeTab
                ? `Inga ${STATUS_LABELS[activeTab as InvoiceStatus]?.toLowerCase() || ""} fakturor`
                : "Inga fakturor skapade än"}
            </p>
            {!activeTab && (
              <Link
                href="/invoices/new"
                className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Skapa din första faktura
              </Link>
            )}
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {invoices.map((inv) => (
              <Link
                key={inv.invoice_id}
                href={`/invoices/${inv.invoice_id}`}
                className="block p-4 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center justify-between mb-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-blue-600">#{inv.invoice_number}</span>
                    <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[inv.status as InvoiceStatus] || ""}`}>
                      {STATUS_LABELS[inv.status as InvoiceStatus] || inv.status}
                    </span>
                  </div>
                  <span className="text-sm font-medium text-gray-900">{formatSEK(inv.total)} kr</span>
                </div>
                <p className="text-sm text-gray-900 truncate">
                  {customers[inv.customer_id]?.name || `Kund #${inv.customer_id}`}
                </p>
                <div className="flex items-center gap-3 mt-1 text-xs text-gray-500">
                  <span>{new Date(inv.invoice_date).toLocaleDateString("sv-SE")}</span>
                  <span>Förfaller: {new Date(inv.due_date).toLocaleDateString("sv-SE")}</span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>

      <div className="mt-6 text-sm text-gray-600">
        Visar {invoices.length} fakturor
      </div>
    </DashboardLayout>
  );
}
