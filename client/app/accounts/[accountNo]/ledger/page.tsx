"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { accountsApi } from "@/lib/api/accounts";
import { vouchersApi } from "@/lib/api/vouchers";
import { Account, LedgerEntry } from "@/types";
import { formatSEK, parseMoney } from "@/lib/money";
import Link from "next/link";

export default function AccountLedgerPage() {
  const params = useParams();
  const router = useRouter();
  const accountNo = parseInt(params.accountNo as string);

  const [account, setAccount] = useState<Account | null>(null);
  const [entries, setEntries] = useState<LedgerEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [period, setPeriod] = useState(""); // Empty string = show all periods
  const [availablePeriods, setAvailablePeriods] = useState<string[]>([]);

  // Fetch available periods on mount
  useEffect(() => {
    const fetchPeriods = async () => {
      try {
        const periods = await vouchersApi.getAllPeriods();
        setAvailablePeriods(periods);
      } catch (error) {
        console.error("Failed to fetch periods:", error);
      }
    };
    fetchPeriods();
  }, []);

  // Fetch ledger data when account or period changes
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [accountData, ledgerData] = await Promise.all([
          accountsApi.getByAccountNo(accountNo),
          accountsApi.getLedger(accountNo, period),
        ]);
        setAccount(accountData);
        setEntries(ledgerData);
      } catch (error) {
        console.error("Failed to fetch ledger data:", error);
      } finally {
        setLoading(false);
      }
    };

    if (!isNaN(accountNo)) {
      fetchData();
    }
  }, [accountNo, period]);

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  if (!account) {
    return (
      <DashboardLayout>
        <div className="text-center py-12">
          <p className="text-gray-500 mb-4">Konto hittades inte</p>
          <Link
            href="/accounts"
            className="text-blue-600 hover:text-blue-700 font-medium"
          >
            ← Tillbaka till kontoplanen
          </Link>
        </div>
      </DashboardLayout>
    );
  }

  // Derive opening balance from the first row: reverse the effect of its first line
  // (running balance in row 0 = opening + debit0 - credit0  =>  opening = balance0 - debit0 + credit0)
  const openingBalance = entries.length > 0
    ? parseMoney(entries[0].balance) - parseMoney(entries[0].debit_amount) + parseMoney(entries[0].credit_amount)
    : 0;
  const closingBalance = entries.length > 0 ? parseMoney(entries[entries.length - 1].balance) : 0;

  return (
    <DashboardLayout>
      {/* Header */}
      <div className="mb-6">
        <Link
          href="/accounts"
          className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block"
        >
          ← Tillbaka till kontoplanen
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">
          Kontoutdrag: {account.account_no} - {account.account_name}
        </h1>
        <div className="mt-2 flex gap-4 text-sm text-gray-600">
          <span>Typ: {account.type === "BS" ? "Balansräkning" : "Resultaträkning"}</span>
          <span>•</span>
          <span>Standardsida: {account.standard_side === "Debit" ? "Debet" : "Kredit"}</span>
          <span>•</span>
          <span>Moms: {account.tax_standard}</span>
        </div>
      </div>

      {/* Period filter */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="max-w-xs">
          <label htmlFor="period" className="block text-sm font-medium text-gray-700 mb-2">
            Period
          </label>
          <select
            id="period"
            value={period}
            onChange={(e) => setPeriod(e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
          >
            <option value="">Alla perioder</option>
            {availablePeriods.map((p) => (
              <option key={p} value={p}>
                {new Date(p + "-01").toLocaleDateString("sv-SE", {
                  year: "numeric",
                  month: "long",
                })}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Ledger table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {entries.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">Inga transaktioner för denna period</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Datum</th>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Ver.nr</th>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Beskrivning</th>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Referens</th>
                  <th className="text-right py-3 px-6 text-sm font-semibold text-gray-700">Debet</th>
                  <th className="text-right py-3 px-6 text-sm font-semibold text-gray-700">Kredit</th>
                  <th className="text-right py-3 px-6 text-sm font-semibold text-gray-700">Saldo</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {/* Opening balance row */}
                <tr className="bg-blue-50 font-medium">
                  <td className="py-3 px-6 text-sm text-gray-900" colSpan={6}>
                    Ingående balans
                  </td>
                  <td className="py-3 px-6 text-sm text-gray-900 text-right">
                    {formatSEK(openingBalance)} kr
                  </td>
                </tr>

                {/* Transaction rows */}
                {entries.map((entry) => (
                  <tr key={`${entry.voucher_id}-${entry.date}`} className="hover:bg-gray-50">
                    <td className="py-3 px-6 text-sm text-gray-600">
                      {new Date(entry.date).toLocaleDateString("sv-SE")}
                    </td>
                    <td className="py-3 px-6 text-sm font-semibold text-blue-600">
                      <Link
                        href={`/vouchers/${entry.voucher_id}`}
                        className="hover:underline"
                      >
                        #{entry.voucher_number}
                      </Link>
                    </td>
                    <td className="py-3 px-6 text-sm text-gray-900">{entry.description}</td>
                    <td className="py-3 px-6 text-sm text-gray-600">{entry.reference}</td>
                    <td className="py-3 px-6 text-sm text-gray-900 text-right">
                      {parseMoney(entry.debit_amount) > 0 && <>{formatSEK(entry.debit_amount)} kr</>}
                    </td>
                    <td className="py-3 px-6 text-sm text-gray-900 text-right">
                      {parseMoney(entry.credit_amount) > 0 && <>{formatSEK(entry.credit_amount)} kr</>}
                    </td>
                    <td className="py-3 px-6 text-sm text-gray-900 text-right font-medium">
                      {formatSEK(entry.balance)} kr
                    </td>
                  </tr>
                ))}

                {/* Closing balance row */}
                <tr className="bg-blue-50 font-medium">
                  <td className="py-3 px-6 text-sm text-gray-900" colSpan={6}>
                    Utgående balans
                  </td>
                  <td className="py-3 px-6 text-sm text-gray-900 text-right">
                    {formatSEK(closingBalance)} kr
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Summary */}
      {entries.length > 0 && (
        <div className="mt-6 bg-gray-50 rounded-lg p-4 flex justify-between text-sm">
          <div>
            <span className="text-gray-600">Antal transaktioner:</span>{" "}
            <span className="font-medium text-gray-900">{entries.length}</span>
          </div>
          <div>
            <span className="text-gray-600">Period:</span>{" "}
            <span className="font-medium text-gray-900">
              {period === ""
                ? "Alla perioder"
                : new Date(period + "-01").toLocaleDateString("sv-SE", {
                    year: "numeric",
                    month: "long",
                  })
              }
            </span>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
