"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { accountsApi } from "@/lib/api/accounts";
import { TaxStandard, AccountType, StandardSide } from "@/types";

export default function NewAccountPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [formData, setFormData] = useState<{
    account_no: string;
    account_name: string;
    account_group: number;
    tax_standard: TaxStandard;
    type: AccountType;
    standard_side: StandardSide;
  }>({
    account_no: "",
    account_name: "",
    account_group: 1,
    tax_standard: "0%",
    type: "P&L",
    standard_side: "Debit",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await accountsApi.create({
        ...formData,
        account_no: parseInt(formData.account_no),
      });
      router.push("/accounts");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create account");
    } finally {
      setLoading(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="max-w-2xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => router.back()}
            className="text-sm text-gray-600 hover:text-gray-900 mb-4 flex items-center gap-2"
          >
            ← Tillbaka
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Nytt konto</h1>
          <p className="text-gray-600 mt-2">Lägg till ett nytt konto i kontoplanen</p>
        </div>

        {/* Form */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Account Number */}
            <div>
              <label htmlFor="account_no" className="block text-sm font-medium text-gray-700 mb-2">
                Kontonummer *
              </label>
              <input
                id="account_no"
                type="number"
                required
                value={formData.account_no}
                onChange={(e) => setFormData({ ...formData, account_no: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
                placeholder="t.ex. 1510"
              />
              <p className="mt-1 text-sm text-gray-500">BAS-kontonummer (4 siffror)</p>
            </div>

            {/* Account Name */}
            <div>
              <label htmlFor="account_name" className="block text-sm font-medium text-gray-700 mb-2">
                Kontonamn *
              </label>
              <input
                id="account_name"
                type="text"
                required
                value={formData.account_name}
                onChange={(e) => setFormData({ ...formData, account_name: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
                placeholder="t.ex. Kundfordringar"
              />
            </div>

            {/* Account Group */}
            <div>
              <label htmlFor="account_group" className="block text-sm font-medium text-gray-700 mb-2">
                Kontogrupp *
              </label>
              <select
                id="account_group"
                required
                value={formData.account_group}
                onChange={(e) => setFormData({ ...formData, account_group: parseInt(e.target.value) })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
              >
                <option value={1}>1 - Tillgångar</option>
                <option value={2}>2 - Eget kapital och skulder</option>
                <option value={3}>3 - Inkomster</option>
                <option value={4}>4 - Kostnader för varor och material</option>
                <option value={5}>5 - Övriga externa kostnader</option>
                <option value={6}>6 - Personalkostnader</option>
                <option value={7}>7 - Avskrivningar och nedskrivningar</option>
                <option value={8}>8 - Finansiella poster</option>
              </select>
              <p className="mt-1 text-sm text-gray-500">BAS-kontogrupp enligt svensk standard</p>
            </div>

            {/* Tax Standard */}
            <div>
              <label htmlFor="tax_standard" className="block text-sm font-medium text-gray-700 mb-2">
                Momssats *
              </label>
              <select
                id="tax_standard"
                required
                value={formData.tax_standard}
                onChange={(e) => setFormData({ ...formData, tax_standard: e.target.value as TaxStandard })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
              >
                <option value="0%">0% - Momsfri</option>
                <option value="6%">6% - Reducerad moms</option>
                <option value="12%">12% - Mellanmoms</option>
                <option value="25%">25% - Normal moms</option>
              </select>
            </div>

            {/* Account Type */}
            <div>
              <label htmlFor="type" className="block text-sm font-medium text-gray-700 mb-2">
                Kontotyp *
              </label>
              <select
                id="type"
                required
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value as AccountType })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
              >
                <option value="BS">BS - Balansräkning</option>
                <option value="P&L">P&L - Resultaträkning</option>
              </select>
            </div>

            {/* Standard Side */}
            <div>
              <label htmlFor="standard_side" className="block text-sm font-medium text-gray-700 mb-2">
                Standardsida *
              </label>
              <select
                id="standard_side"
                required
                value={formData.standard_side}
                onChange={(e) => setFormData({ ...formData, standard_side: e.target.value as StandardSide })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
              >
                <option value="Debit">Debet</option>
                <option value="Credit">Kredit</option>
              </select>
              <p className="mt-1 text-sm text-gray-500">
                Vanlig sida för ökningar på detta konto
              </p>
            </div>

            {/* Buttons */}
            <div className="flex gap-4 pt-4">
              <button
                type="button"
                onClick={() => router.back()}
                className="flex-1 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium"
              >
                Avbryt
              </button>
              <button
                type="submit"
                disabled={loading}
                className="flex-1 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? "Sparar..." : "Skapa konto"}
              </button>
            </div>
          </form>
        </div>

        {/* Help Text */}
        <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="text-sm font-semibold text-blue-900 mb-2">💡 Tips</h3>
          <ul className="text-sm text-blue-700 space-y-1">
            <li>• BAS-kontonummer följer svensk standard (1000-8999)</li>
            <li>• Kontogrupp 1-2 är för balansräkning (BS)</li>
            <li>• Kontogrupp 3-8 är för resultaträkning (IS)</li>
            <li>• Debetsida används för tillgångar och kostnader</li>
            <li>• Kreditsida används för skulder, eget kapital och intäkter</li>
          </ul>
        </div>
      </div>
    </DashboardLayout>
  );
}
