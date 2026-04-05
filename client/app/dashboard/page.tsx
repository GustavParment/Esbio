"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { vouchersApi } from "@/lib/api/vouchers";
import { accountsApi } from "@/lib/api/accounts";
import { Voucher, Account } from "@/types";
import { useAuth } from "@/lib/contexts/AuthContext";
import Link from "next/link";

export default function DashboardPage() {
  const { user } = useAuth();
  const [vouchers, setVouchers] = useState<Voucher[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPeriod] = useState(() => {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
  });

  useEffect(() => {
    const fetchData = async () => {
      if (!user) {
        setLoading(false);
        return;
      }

      try {
        // Only fetch user's own vouchers (or all if Admin)
        const vouchersPromise = user.role === "Admin"
          ? vouchersApi.getAll()
          : vouchersApi.getByUser(user.user_id);

        const [vouchersData, accountsData] = await Promise.all([
          vouchersPromise,
          accountsApi.getAll(),
        ]);
        // Handle null/undefined responses by defaulting to empty arrays
        const vouchersArray = Array.isArray(vouchersData) ? vouchersData : [];
        // Filter out corrected vouchers and take latest 5
        const activeVouchers = vouchersArray.filter(v => !v.corrected_by_voucher_id);
        setVouchers(activeVouchers.slice(0, 5)); // Latest 5 active vouchers
        setAccounts(Array.isArray(accountsData) ? accountsData : []);
      } catch (error) {
        console.error("Failed to fetch dashboard data:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [user?.user_id, user?.role]);

  const stats = [
    { name: "Verifikat denna månad", value: vouchers.length, color: "bg-blue-500",
      svg: <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg> },
    { name: "Totalt konton", value: accounts.length, color: "bg-blue-500",
      svg: <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg> },
    { name: "Aktuell period", value: currentPeriod, color: "bg-blue-500",
      svg: <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg> },
    { name: "Status", value: "Aktiv", color: "bg-blue-500",
      svg: <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg> },
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
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 mt-2">Översikt av din bokföring</p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {stats.map((stat) => (
          <div key={stat.name} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">{stat.name}</p>
                <p className="text-2xl font-bold text-gray-900 mt-2">{stat.value}</p>
              </div>
              <div className={`${stat.color} w-12 h-12 rounded-lg flex items-center justify-center`}>
                {stat.svg}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Recent vouchers */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-bold text-gray-900">Senaste verifikaten</h2>
          <Link
            href="/vouchers"
            className="text-sm text-blue-600 hover:text-blue-700 font-medium"
          >
            Visa alla →
          </Link>
        </div>

        {vouchers.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 mb-4">Inga verifikat hittades</p>
            <Link
              href="/vouchers/new"
              className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Skapa nytt verifikat
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Nr</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Datum</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Beskrivning</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Referens</th>
                  <th className="text-right py-3 px-4 text-sm font-semibold text-gray-700">Belopp</th>
                </tr>
              </thead>
              <tbody>
                {vouchers.map((voucher) => (
                  <tr key={voucher.voucher_id} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="py-3 px-4 text-sm font-semibold text-blue-600">#{voucher.voucher_number}</td>
                    <td className="py-3 px-4 text-sm text-gray-600">
                      {new Date(voucher.date).toLocaleDateString("sv-SE")}
                    </td>
                    <td className="py-3 px-4 text-sm text-gray-900">{voucher.description}</td>
                    <td className="py-3 px-4 text-sm text-gray-600">{voucher.reference}</td>
                    <td className="py-3 px-4 text-sm text-gray-900 text-right font-medium">
                      {voucher.total_amount.toLocaleString("sv-SE")} kr
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Quick actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Link
          href="/vouchers/new"
          className="rounded-xl shadow-lg p-6 text-white hover:shadow-xl transition-shadow"
          style={{ background: "linear-gradient(135deg, #2D9F83, #238B72)" }}
        >
          <div className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </div>
          <h3 className="text-lg font-semibold mb-2">Nytt verifikat</h3>
          <p className="text-sm opacity-80">Skapa ett nytt bokföringsverifikat</p>
        </Link>

        <Link
          href="/accounts"
          className="rounded-xl shadow-lg p-6 text-white hover:shadow-xl transition-shadow"
          style={{ background: "linear-gradient(135deg, #1A6B5A, #145248)" }}
        >
          <div className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
          </div>
          <h3 className="text-lg font-semibold mb-2">Kontoplanen</h3>
          <p className="text-sm opacity-80">Visa och hantera kontoplanen</p>
        </Link>

        <Link
          href="/reports"
          className="rounded-xl shadow-lg p-6 text-white hover:shadow-xl transition-shadow"
          style={{ background: "linear-gradient(135deg, #0F4D40, #0B3B31)" }}
        >
          <div className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
          </div>
          <h3 className="text-lg font-semibold mb-2">Rapporter</h3>
          <p className="text-sm opacity-80">Visa finansiella rapporter</p>
        </Link>
      </div>
    </DashboardLayout>
  );
}
