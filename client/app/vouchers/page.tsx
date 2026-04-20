"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { vouchersApi } from "@/lib/api/vouchers";
import { Voucher } from "@/types";
import { formatSEK, sumMoney } from "@/lib/money";
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

      // Fetch vouchers for the selected company (scoped by cookie)
      const companyVouchers = await vouchersApi.getByCompany();
      data = currentPeriod
        ? companyVouchers.filter(v => v.period === currentPeriod)
        : companyVouchers;

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
      <div className="mb-8">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Verifikat</h1>
            <p className="text-gray-600 mt-2">Hantera bokföringsverifikat</p>
          </div>
          <div className="flex gap-3">
            <Link
              href="/vouchers/scan"
              className="px-4 py-2.5 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors font-medium text-sm text-center"
            >
              Skanna kvitto / PDF
            </Link>
            <Link
              href="/vouchers/new"
              className="px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm text-center"
            >
              + Nytt verifikat
            </Link>
          </div>
        </div>
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
          <div className="divide-y divide-gray-100">
            {filteredVouchers.map((voucher) => (
              <Link
                key={voucher.voucher_id}
                href={`/vouchers/${voucher.voucher_id}`}
                className={`block p-4 hover:bg-gray-50 transition-colors ${voucher.corrected_by_voucher_id ? 'bg-red-50 opacity-60' : ''}`}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-semibold text-blue-600">
                    #{voucher.voucher_number}
                    {voucher.corrected_by_voucher_id && (
                      <span className="ml-2 text-xs text-red-600 font-normal">(rättad)</span>
                    )}
                  </span>
                  <span className="text-sm font-medium text-gray-900">
                    {formatSEK(voucher.total_amount)} kr
                  </span>
                </div>
                <p className="text-sm text-gray-900 truncate">{voucher.description}</p>
                <div className="flex items-center gap-3 mt-1 text-xs text-gray-500">
                  <span>{new Date(voucher.date).toLocaleDateString("sv-SE")}</span>
                  {voucher.reference && <span>{voucher.reference}</span>}
                  <span>
                    {new Date(voucher.period + "-01").toLocaleDateString("sv-SE", {
                      year: "numeric",
                      month: "short",
                    })}
                  </span>
                </div>
              </Link>
            ))}
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
          {formatSEK(
            sumMoney(
              filteredVouchers.filter((v) => !v.corrected_by_voucher_id),
              (v) => v.total_amount
            )
          )}{" "}
          kr
        </div>
      </div>
    </DashboardLayout>
  );
}
