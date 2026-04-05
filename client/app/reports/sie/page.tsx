"use client";

import { useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { reportsApi } from "@/lib/api/reports";
import Link from "next/link";

export default function SIEExportPage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const currentYear = new Date().getFullYear();
  const [fromDate, setFromDate] = useState(`${currentYear}-01-01`);
  const [toDate, setToDate] = useState(`${currentYear}-12-31`);

  const handleExport = async () => {
    if (!fromDate || !toDate) {
      setError("Både från- och till-datum krävs");
      return;
    }

    setLoading(true);
    setError("");
    setSuccess("");
    try {
      await reportsApi.downloadSIE(fromDate, toDate);
      setSuccess("SIE4-fil exporterad!");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte exportera SIE-fil. Kontrollera att företagsnamn är ifyllt under Inställningar.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="mb-8">
        <Link
          href="/reports"
          className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block"
        >
          &larr; Tillbaka till rapporter
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">SIE4-export</h1>
        <p className="text-gray-600 mt-2">
          Exportera din bokföring i SIE4-format för Skatteverket eller din revisor
        </p>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label htmlFor="from_date" className="block text-sm font-medium text-gray-700 mb-2">
              Från datum (räkenskapsår start)
            </label>
            <input
              id="from_date"
              type="date"
              value={fromDate}
              onChange={(e) => setFromDate(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
            />
          </div>

          <div>
            <label htmlFor="to_date" className="block text-sm font-medium text-gray-700 mb-2">
              Till datum (räkenskapsår slut)
            </label>
            <input
              id="to_date"
              type="date"
              value={toDate}
              onChange={(e) => setToDate(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={handleExport}
              disabled={loading}
              className="w-full px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50"
            >
              {loading ? "Exporterar..." : "Ladda ner SIE4-fil"}
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-700">{error}</p>
        </div>
      )}

      {success && (
        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <p className="text-green-700">{success}</p>
        </div>
      )}

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-3">Om SIE4-export</h3>
        <ul className="space-y-2 text-sm text-gray-600">
          <li>SIE4 innehåller hela din bokföring: kontoplan, ingående/utgående balanser och alla verifikat</li>
          <li>Filen kan importeras i andra bokföringsprogram eller skickas till din revisor</li>
          <li>Se till att företagsnamn och organisationsnummer är ifyllt under <Link href="/settings" className="text-blue-600 hover:underline">Inställningar</Link></li>
          <li>Formatet följer SIE-standarden och använder BAS2024 kontoplan</li>
        </ul>
      </div>
    </DashboardLayout>
  );
}
