"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { reportsApi, VATReport } from "@/lib/api/reports";
import { parseMoney } from "@/lib/money";
import Link from "next/link";

export default function VATReportPage() {
  const [report, setReport] = useState<VATReport | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [downloading, setDownloading] = useState(false);

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

  const downloadDeclaration = async () => {
    if (!fromDate || !toDate) return;
    setDownloading(true);
    setError("");
    try {
      await reportsApi.downloadVATDeclaration(fromDate, toDate);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte ladda ner momsdeklaration");
    } finally {
      setDownloading(false);
    }
  };

  const formatCurrency = (value: string | number) =>
    parseMoney(value).toLocaleString("sv-SE", {
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
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
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
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={fetchReport}
              disabled={loading}
              className="w-full px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50"
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
          <div className="p-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            <h2 className="text-xl font-bold text-gray-900">
              Period: {new Date(report.period.from_date).toLocaleDateString("sv-SE")} till{" "}
              {new Date(report.period.to_date).toLocaleDateString("sv-SE")}
            </h2>
            <button
              onClick={downloadDeclaration}
              disabled={downloading}
              className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors font-medium text-sm disabled:opacity-50 whitespace-nowrap"
            >
              {downloading ? "Laddar ner..." : "Ladda ner momsdeklaration (PDF)"}
            </button>
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

          {/* Output VAT Totals */}
          <div className="px-6 py-4 bg-gray-50 border-t-2 border-gray-200">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-600">Total försäljning</p>
                <p className="text-xl font-bold text-gray-900">
                  {formatCurrency(report.total_sales)} kr
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm text-gray-600">Utgående moms</p>
                <p className="text-xl font-bold text-red-600">
                  {formatCurrency(report.total_vat)} kr
                </p>
              </div>
            </div>
          </div>

          {/* Input VAT Section */}
          <div className="px-6 py-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4 mt-2">INGÅENDE MOMS</h3>
            {report.input_entries && report.input_entries.length > 0 ? (
              <table className="w-full">
                <thead>
                  <tr className="border-b-2 border-gray-200">
                    <th className="py-3 text-left text-sm font-semibold text-gray-700">Konto</th>
                    <th className="py-3 text-left text-sm font-semibold text-gray-700">Kontonamn</th>
                    <th className="py-3 text-right text-sm font-semibold text-gray-700">Belopp</th>
                  </tr>
                </thead>
                <tbody>
                  {report.input_entries.map((entry) => (
                    <tr key={entry.account_no} className="border-b border-gray-100">
                      <td className="py-3 text-sm text-gray-900 font-medium">{entry.account_no}</td>
                      <td className="py-3 text-sm text-gray-900">{entry.account_name}</td>
                      <td className="py-3 text-sm text-gray-900 text-right">
                        {formatCurrency(entry.amount)} kr
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr className="border-t-2 border-gray-300">
                    <td colSpan={2} className="py-3 text-sm font-bold text-gray-900">
                      Total ingående moms
                    </td>
                    <td className="py-3 text-sm font-bold text-green-600 text-right">
                      {formatCurrency(report.total_input_vat)} kr
                    </td>
                  </tr>
                </tfoot>
              </table>
            ) : (
              <p className="text-gray-500 italic">Ingen ingående moms för denna period</p>
            )}
          </div>

          {/* Net VAT */}
          {(() => {
            const netVatNum = parseMoney(report.net_vat);
            const isPayable = netVatNum >= 0;
            return (
              <div className={`px-6 py-6 border-t-4 ${isPayable ? "bg-red-50 border-red-500" : "bg-green-50 border-green-500"}`}>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <p className="text-sm text-gray-600">Utgående moms</p>
                    <p className="text-lg font-bold text-red-600">
                      {formatCurrency(report.total_vat)} kr
                    </p>
                  </div>
                  <div className="text-center">
                    <p className="text-sm text-gray-600">Ingående moms</p>
                    <p className="text-lg font-bold text-green-600">
                      -{formatCurrency(report.total_input_vat)} kr
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-gray-600">
                      {isPayable ? "Moms att betala" : "Moms att få tillbaka"}
                    </p>
                    <p className={`text-2xl font-bold ${isPayable ? "text-red-600" : "text-green-600"}`}>
                      {formatCurrency(Math.abs(netVatNum))} kr
                    </p>
                  </div>
                </div>
              </div>
            );
          })()}
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
