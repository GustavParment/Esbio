"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { reportsApi, VATReport } from "@/lib/api/reports";
import Link from "next/link";

export default function VATReportPage() {
  const [report, setReport] = useState<VATReport | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Default to current year
  const currentYear = new Date().getFullYear();
  const [fromDate, setFromDate] = useState(`${currentYear}-01-01`);
  const [toDate, setToDate] = useState(`${currentYear}-12-31`);

  const fetchReport = async () => {
    if (!fromDate || !toDate) {
      setError("Både från- och till-datum krävs");
      return;
    }

    setLoading(true);
    setError("");
    try {
      const data = await reportsApi.getVATReport(fromDate, toDate);
      setReport(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte hämta momsrapport");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReport();
  }, []);

  const formatCurrency = (value: number) =>
    value.toLocaleString("sv-SE", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });

  return (
    <DashboardLayout>
      {/* Header */}
      <div className="mb-8">
        <Link
          href="/reports"
          className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block"
        >
          &larr; Tillbaka till rapporter
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">Momsrapport</h1>
        <p className="text-gray-600 mt-2">Visa momsunderlag och beräknad moms per momssats</p>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label htmlFor="from_date" className="block text-sm font-medium text-gray-700 mb-2">
              Från datum
            </label>
            <input
              id="from_date"
              type="date"
              value={fromDate}
              onChange={(e) => setFromDate(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent text-gray-900"
            />
          </div>

          <div>
            <label htmlFor="to_date" className="block text-sm font-medium text-gray-700 mb-2">
              Till datum
            </label>
            <input
              id="to_date"
              type="date"
              value={toDate}
              onChange={(e) => setToDate(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent text-gray-900"
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={fetchReport}
              disabled={loading}
              className="w-full px-6 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium disabled:opacity-50"
            >
              {loading ? "Laddar..." : "Visa momsrapport"}
            </button>
          </div>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-700">{error}</p>
        </div>
      )}

      {/* Results */}
      {report && !loading && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="p-6">
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Period: {new Date(report.period.from_date).toLocaleDateString("sv-SE")} till{" "}
              {new Date(report.period.to_date).toLocaleDateString("sv-SE")}
            </h2>
          </div>

          {/* VAT Entries Table */}
          <div className="px-6 pb-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">MOMSUNDERLAG PER MOMSSATS</h3>
            {report.entries.length === 0 ? (
              <p className="text-gray-500 italic">Ingen försäljning med moms för denna period</p>
            ) : (
              <table className="w-full">
                <thead>
                  <tr className="border-b-2 border-gray-200">
                    <th className="py-3 text-left text-sm font-semibold text-gray-700">Momssats</th>
                    <th className="py-3 text-right text-sm font-semibold text-gray-700">
                      Försäljning exkl. moms
                    </th>
                    <th className="py-3 text-right text-sm font-semibold text-gray-700">
                      Beräknad moms
                    </th>
                    <th className="py-3 text-right text-sm font-semibold text-gray-700">
                      Försäljning inkl. moms
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {report.entries.map((entry) => (
                    <tr key={entry.tax_code} className="border-b border-gray-100">
                      <td className="py-3 text-sm text-gray-900 font-medium">{entry.tax_rate}</td>
                      <td className="py-3 text-sm text-gray-900 text-right">
                        {formatCurrency(entry.total_sales)} kr
                      </td>
                      <td className="py-3 text-sm text-gray-900 text-right">
                        {formatCurrency(entry.total_vat)} kr
                      </td>
                      <td className="py-3 text-sm text-gray-900 text-right">
                        {formatCurrency(entry.total_sales + entry.total_vat)} kr
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {/* Totals */}
          <div className="px-6 py-6 bg-red-50 border-t-4 border-red-500">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-600">Total försäljning</p>
                <p className="text-xl font-bold text-gray-900">
                  {formatCurrency(report.total_sales)} kr
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm text-gray-600">Total utgående moms</p>
                <p className="text-2xl font-bold text-red-600">
                  {formatCurrency(report.total_vat)} kr
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!report && !loading && !error && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
          <p className="text-gray-500">
            Välj en period och klicka på &quot;Visa momsrapport&quot;
          </p>
        </div>
      )}
    </DashboardLayout>
  );
}
