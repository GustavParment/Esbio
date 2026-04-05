"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { vouchersApi } from "@/lib/api/vouchers";
import { Voucher } from "@/types";
import { useAuth } from "@/lib/contexts/AuthContext";
import Link from "next/link";

export default function VouchersPage() {
  const { user } = useAuth();
  const [vouchers, setVouchers] = useState<Voucher[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPeriod, setCurrentPeriod] = useState("");
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

  useEffect(() => {
    if (user) {
      fetchVouchers();
    }
  }, [currentPeriod, user]);

  const fetchVouchers = async () => {
    if (!user) return;

    setLoading(true);
    try {
      let data: Voucher[];

      // Admin can see all vouchers, others only their own
      if (user.role === "Admin") {
        data = currentPeriod
          ? await vouchersApi.getByPeriod(currentPeriod)
          : await vouchersApi.getAll();
      } else {
        // For non-admin users, get their vouchers and filter by period
        const allUserVouchers = await vouchersApi.getByUser(user.user_id);
        data = currentPeriod
          ? allUserVouchers.filter(v => v.period === currentPeriod)
          : allUserVouchers;
      }

      setVouchers(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Failed to fetch vouchers:", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredVouchers = vouchers.filter((voucher) => {
    if (searchTerm === "") return true;
    return (
      voucher.voucher_id.toString().includes(searchTerm) ||
      voucher.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
      voucher.reference.toLowerCase().includes(searchTerm.toLowerCase())
    );
  });


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
          <h1 className="text-3xl font-bold text-gray-900">Verifikat</h1>
          <p className="text-gray-600 mt-2">Hantera bokföringsverifikat</p>
        </div>
        <Link
          href="/vouchers/new"
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
        >
          + Nytt verifikat
        </Link>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-2">
              Sök verifikat
            </label>
            <input
              id="search"
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Verifikat-ID, beskrivning eller referens..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
            />
          </div>

          <div>
            <label htmlFor="period" className="block text-sm font-medium text-gray-700 mb-2">
              Period
            </label>
            <select
              id="period"
              value={currentPeriod}
              onChange={(e) => setCurrentPeriod(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
            >
              <option value="">Alla perioder</option>
              {availablePeriods.map((period) => (
                <option key={period} value={period}>
                  {new Date(period + "-01").toLocaleDateString("sv-SE", {
                    year: "numeric",
                    month: "long",
                  })}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Vouchers table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {filteredVouchers.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 mb-4">Inga verifikat hittades för denna period</p>
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
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Ver.nr</th>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Datum</th>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Beskrivning</th>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Referens</th>
                  <th className="text-left py-3 px-6 text-sm font-semibold text-gray-700">Period</th>
                  <th className="text-right py-3 px-6 text-sm font-semibold text-gray-700">Belopp</th>
                  <th className="text-right py-3 px-6 text-sm font-semibold text-gray-700">Åtgärder</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filteredVouchers.map((voucher) => (
                  <tr
                    key={voucher.voucher_id}
                    className={`hover:bg-gray-50 ${voucher.corrected_by_voucher_id ? 'bg-red-50 opacity-60' : ''}`}
                  >
                    <td className="py-4 px-6 text-sm font-semibold text-blue-600">
                      #{voucher.voucher_number}
                      {voucher.corrected_by_voucher_id && (
                        <span className="ml-2 text-xs text-red-600 font-normal">(rättad)</span>
                      )}
                    </td>
                    <td className="py-4 px-6 text-sm text-gray-600">
                      {new Date(voucher.date).toLocaleDateString("sv-SE")}
                    </td>
                    <td className="py-4 px-6 text-sm text-gray-900">{voucher.description}</td>
                    <td className="py-4 px-6 text-sm text-gray-600">{voucher.reference}</td>
                    <td className="py-4 px-6 text-sm text-gray-600">
                      {new Date(voucher.period + "-01").toLocaleDateString("sv-SE", {
                        year: "numeric",
                        month: "short",
                      })}
                    </td>
                    <td className="py-4 px-6 text-sm text-gray-900 text-right font-medium">
                      {voucher.total_amount.toLocaleString("sv-SE", {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })}{" "}
                      kr
                    </td>
                    <td className="py-4 px-6 text-sm text-right">
                      <Link
                        href={`/vouchers/${voucher.voucher_id}`}
                        className="text-blue-600 hover:text-blue-700 font-medium"
                      >
                        Visa →
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Summary */}
      <div className="mt-6 flex items-center justify-between text-sm text-gray-600">
        <div>
          Visar {filteredVouchers.length} av {vouchers.length} verifikat
        </div>
        <div className="font-medium">
          Total summa:{" "}
          {filteredVouchers
            .filter((v) => !v.corrected_by_voucher_id) // Exclude corrected vouchers
            .reduce((sum, v) => sum + v.total_amount, 0)
            .toLocaleString("sv-SE", {
              minimumFractionDigits: 2,
              maximumFractionDigits: 2,
            })}{" "}
          kr
        </div>
      </div>
    </DashboardLayout>
  );
}
