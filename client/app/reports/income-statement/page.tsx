"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { reportsApi, IncomeStatement } from "@/lib/api/reports";
import { formatSEK, parseMoney } from "@/lib/money";
import Link from "next/link";

export default function IncomeStatementPage() {
  const [statement, setStatement] = useState<IncomeStatement | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Default to current year
  const currentYear = new Date().getFullYear();
  const [fromDate, setFromDate] = useState(`${currentYear}-01-01`);
  const [toDate, setToDate] = useState(`${currentYear}-12-31`);

  const fetchStatement = async () => {
    if (!fromDate || !toDate) {
      setError("Både från- och till-datum krävs");
      return;
    }

    setLoading(true);
    setError("");
    try {
      const data = await reportsApi.getIncomeStatement(fromDate, toDate);
      setStatement(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte hämta resultaträkning");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatement();
  }, []);

  return (
    <DashboardLayout>
      {/* Header */}
      <div className="mb-8">
        <Link
          href="/reports"
          className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block"
        >
          ← Tillbaka till rapporter
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">Resultaträkning</h1>
        <p className="text-gray-600 mt-2">Visa intäkter och kostnader för en period</p>
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
              onClick={fetchStatement}
              disabled={loading}
              className="w-full px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50"
            >
              {loading ? "Laddar..." : "Visa resultaträkning"}
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
      {statement && !loading && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="p-6">
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Period: {new Date(statement.period.from_date).toLocaleDateString("sv-SE")} till{" "}
              {new Date(statement.period.to_date).toLocaleDateString("sv-SE")}
            </h2>
          </div>

          {/* Income Section */}
          <div className="px-6 pb-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">INTÄKTER</h3>
            {statement.income.length === 0 ? (
              <p className="text-gray-500 italic">Inga intäkter för denna period</p>
            ) : (
              <table className="w-full">
                <tbody>
                  {statement.income.map((entry) => (
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
                        {formatSEK(entry.balance)} kr
                      </td>
                    </tr>
                  ))}
                  <tr className="border-t-2 border-gray-300">
                    <td className="py-3 text-sm font-bold text-gray-900">Summa intäkter</td>
                    <td className="py-3 text-sm font-bold text-gray-900 text-right">
                      {formatSEK(statement.total_income)} kr
                    </td>
                  </tr>
                </tbody>
              </table>
            )}
          </div>

          {/* Expenses Section */}
          <div className="px-6 pb-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4 mt-6">KOSTNADER</h3>
            {statement.expenses.length === 0 ? (
              <p className="text-gray-500 italic">Inga kostnader för denna period</p>
            ) : (
              <table className="w-full">
                <tbody>
                  {statement.expenses.map((entry) => (
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
                        {formatSEK(entry.balance)} kr
                      </td>
                    </tr>
                  ))}
                  <tr className="border-t-2 border-gray-300">
                    <td className="py-3 text-sm font-bold text-gray-900">Summa kostnader</td>
                    <td className="py-3 text-sm font-bold text-gray-900 text-right">
                      {formatSEK(statement.total_expenses)} kr
                    </td>
                  </tr>
                </tbody>
              </table>
            )}
          </div>

          {/* Net Result */}
          <div className="px-6 py-6 bg-blue-50 border-t-4 border-blue-500">
            <div className="flex justify-between items-center">
              <h3 className="text-xl font-bold text-gray-900">RESULTAT</h3>
              <p
                className={`text-2xl font-bold ${
                  parseMoney(statement.net_result) >= 0 ? "text-green-600" : "text-red-600"
                }`}
              >
                {formatSEK(statement.net_result)} kr
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!statement && !loading && !error && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
          <p className="text-gray-500">Välj en period och klicka på "Visa resultaträkning"</p>
        </div>
      )}
    </DashboardLayout>
  );
}
