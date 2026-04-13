"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { accountsApi } from "@/lib/api/accounts";
import { Account } from "@/types";
import Link from "next/link";

export default function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<number | "all">("all");
  const [searchTerm, setSearchTerm] = useState("");

  useEffect(() => {
    fetchAccounts();
  }, []);

  const fetchAccounts = async () => {
    try {
      const data = await accountsApi.getAll();
      setAccounts(data);
    } catch (error) {
      console.error("Failed to fetch accounts:", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredAccounts = accounts.filter((account) => {
    const matchesGroup = filter === "all" || account.account_group === filter;
    const matchesSearch =
      searchTerm === "" ||
      account.account_no.toString().includes(searchTerm) ||
      account.account_name.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesGroup && matchesSearch;
  });

  const accountGroups = [
    { value: "all", label: "Alla grupper" },
    { value: 1, label: "1 - Tillgångar" },
    { value: 2, label: "2 - Eget kapital och skulder" },
    { value: 3, label: "3 - Rörelsens intäkter" },
    { value: 4, label: "4 - Rörelsens kostnader" },
    { value: 5, label: "5 - Nedskrivningar" },
    { value: 6, label: "6 - Finansiella poster" },
    { value: 7, label: "7 - Bokslutsdispositioner" },
    { value: 8, label: "8 - Skatter" },
  ];

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
      {/* Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Kontoplan</h1>
          <p className="text-gray-600 mt-2">BAS-kontoplan för bokföring</p>
        </div>
        <Link
          href="/accounts/new"
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
        >
          + Nytt konto
        </Link>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-2">
              Sök konto
            </label>
            <input
              id="search"
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Kontonummer eller namn..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
            />
          </div>

          <div>
            <label htmlFor="group" className="block text-sm font-medium text-gray-700 mb-2">
              Kontogrupp
            </label>
            <select
              id="group"
              value={filter}
              onChange={(e) => setFilter(e.target.value === "all" ? "all" : Number(e.target.value))}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
            >
              {accountGroups.map((group) => (
                <option key={group.value} value={group.value}>
                  {group.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Accounts table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {filteredAccounts.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 mb-4">Inga konton hittades</p>
            <Link
              href="/accounts/new"
              className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Skapa nytt konto
            </Link>
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {filteredAccounts.map((account) => (
              <div key={account.account_no} className="p-4 hover:bg-gray-50">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-gray-900">
                    {account.account_no} — {account.account_name}
                  </span>
                  <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${
                    account.type === "BS" ? "bg-blue-100 text-blue-800" : "bg-green-100 text-green-800"
                  }`}>
                    {account.type}
                  </span>
                </div>
                <div className="flex items-center gap-3 text-xs text-gray-500 mb-2">
                  <span>Grupp {account.account_group}</span>
                  <span>{account.standard_side}</span>
                  <span>{account.tax_standard}</span>
                </div>
                <div className="flex gap-4">
                  <Link href={`/accounts/${account.account_no}/ledger`}
                    className="text-xs text-blue-600 hover:text-blue-700 font-medium">
                    Kontoutdrag
                  </Link>
                  <Link href={`/accounts/${account.account_no}`}
                    className="text-xs text-gray-600 hover:text-gray-700 font-medium">
                    Redigera
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Summary */}
      <div className="mt-6 text-sm text-gray-600">
        Visar {filteredAccounts.length} av {accounts.length} konton
      </div>
    </DashboardLayout>
  );
}
