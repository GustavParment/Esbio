"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { reportsApi, BalanceSheet } from "@/lib/api/reports";
import Link from "next/link";

export default function BalanceSheetPage() {
  const [sheet, setSheet] = useState<BalanceSheet | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Default to today
  const today = new Date().toISOString().split("T")[0];
  const [asOfDate, setAsOfDate] = useState(today);

  const fetchSheet = async () => {
    if (!asOfDate) {
      setError("Datum krävs");
      return;
    }

    setLoading(true);
    setError("");
    try {
      const data = await reportsApi.getBalanceSheet(asOfDate);
      setSheet(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte hämta balansräkning");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSheet();
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
        <h1 className="text-3xl font-bold text-gray-900">Balansräkning</h1>
        <p className="text-gray-600 mt-2">Visa tillgångar, eget kapital och skulder per datum</p>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label htmlFor="as_of_date" className="block text-sm font-medium text-gray-700 mb-2">
              Per datum
            </label>
            <input
              id="as_of_date"
              type="date"
              value={asOfDate}
              onChange={(e) => setAsOfDate(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent text-gray-900"
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={fetchSheet}
              disabled={loading}
              className="w-full px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors font-medium disabled:opacity-50"
            >
              {loading ? "Laddar..." : "Visa balansräkning"}
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
      {sheet && !loading && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="p-6">
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Per datum: {new Date(sheet.as_of_date).toLocaleDateString("sv-SE")}
            </h2>
          </div>

          {/* Assets Section */}
          <div className="px-6 pb-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">TILLGÅNGAR</h3>
            {sheet.assets.length === 0 ? (
              <p className="text-gray-500 italic">Inga tillgångar bokförda</p>
            ) : (
              <table className="w-full">
                <tbody>
                  {sheet.assets.map((entry) => (
                    <tr key={entry.account_no} className="border-b border-gray-100">
                      <td className="py-2 text-sm text-gray-600">
                        <Link
                          href={`/accounts/${entry.account_no}/ledger`}
                          className="hover:text-blue-600"
                        >
                          {entry.account_no} {entry.account_name}
                        </Link>
                      </td>
                      <td className="py-2 text-sm text-gray-900 text-right font-medium">
                        {formatCurrency(entry.balance)} kr
                      </td>
                    </tr>
                  ))}
                  <tr className="border-t-2 border-gray-300">
                    <td className="py-3 text-sm font-bold text-gray-900">Summa tillgångar</td>
                    <td className="py-3 text-sm font-bold text-gray-900 text-right">
                      {formatCurrency(sheet.total_assets)} kr
                    </td>
                  </tr>
                </tbody>
              </table>
            )}
          </div>

          {/* Equity & Liabilities Section */}
          <div className="px-6 pb-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4 mt-6">
              EGET KAPITAL OCH SKULDER
            </h3>
            {sheet.equity_liabilities.length === 0 && sheet.net_result === 0 ? (
              <p className="text-gray-500 italic">Inget eget kapital eller skulder bokförda</p>
            ) : (
              <table className="w-full">
                <tbody>
                  {sheet.equity_liabilities.map((entry) => (
                    <tr key={entry.account_no} className="border-b border-gray-100">
                      <td className="py-2 text-sm text-gray-600">
                        <Link
                          href={`/accounts/${entry.account_no}/ledger`}
                          className="hover:text-blue-600"
                        >
                          {entry.account_no} {entry.account_name}
                        </Link>
                      </td>
                      <td className="py-2 text-sm text-gray-900 text-right font-medium">
                        {formatCurrency(entry.balance)} kr
                      </td>
                    </tr>
                  ))}
                  {sheet.net_result !== 0 && (
                    <tr className="border-b border-gray-100">
                      <td className="py-2 text-sm text-gray-600 italic">
                        Periodens resultat
                      </td>
                      <td className="py-2 text-sm text-gray-900 text-right font-medium">
                        {formatCurrency(sheet.net_result)} kr
                      </td>
                    </tr>
                  )}
                  <tr className="border-t-2 border-gray-300">
                    <td className="py-3 text-sm font-bold text-gray-900">
                      Summa eget kapital och skulder
                    </td>
                    <td className="py-3 text-sm font-bold text-gray-900 text-right">
                      {formatCurrency(sheet.total_equity_liabilities)} kr
                    </td>
                  </tr>
                </tbody>
              </table>
            )}
          </div>

          {/* Balance Check */}
          <div
            className={`px-6 py-6 border-t-4 ${
              Math.abs(sheet.total_assets - sheet.total_equity_liabilities) < 0.01
                ? "bg-green-50 border-green-500"
                : "bg-red-50 border-red-500"
            }`}
          >
            <div className="flex justify-between items-center">
              <h3 className="text-xl font-bold text-gray-900">BALANS</h3>
              <div className="text-right">
                {Math.abs(sheet.total_assets - sheet.total_equity_liabilities) < 0.01 ? (
                  <p className="text-lg font-bold text-green-600">Balanserar</p>
                ) : (
                  <div>
                    <p className="text-lg font-bold text-red-600">Balanserar inte</p>
                    <p className="text-sm text-red-500">
                      Differens:{" "}
                      {formatCurrency(
                        sheet.total_assets - sheet.total_equity_liabilities
                      )}{" "}
                      kr
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!sheet && !loading && !error && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
          <p className="text-gray-500">Välj ett datum och klicka på &quot;Visa balansräkning&quot;</p>
        </div>
      )}
    </DashboardLayout>
  );
}
