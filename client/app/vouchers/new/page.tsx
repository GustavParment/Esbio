"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { vouchersApi } from "@/lib/api/vouchers";
import { accountsApi } from "@/lib/api/accounts";
import { lineItemsApi } from "@/lib/api/lineitems";
import { Account, CreateLineItemRequest } from "@/types";
import { useAuth } from "@/lib/contexts/AuthContext";
import AccountSearch from "@/components/ui/AccountSearch";

interface LineItemForm {
  id: string;
  account_no: string;
  debit_amount: string;
  credit_amount: string;
  tax_code: number;
  description: string;
}

export default function NewVoucherPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [accounts, setAccounts] = useState<Account[]>([]);

  const [formData, setFormData] = useState({
    date: new Date().toISOString().split("T")[0],
    description: "",
    reference: "",
  });

  const [lineItems, setLineItems] = useState<LineItemForm[]>([
    { id: "1", account_no: "", debit_amount: "", credit_amount: "", tax_code: 0, description: "" },
    { id: "2", account_no: "", debit_amount: "", credit_amount: "", tax_code: 0, description: "" },
  ]);

  useEffect(() => {
    const fetchAccounts = async () => {
      try {
        const data = await accountsApi.getAll();
        setAccounts(data || []);
      } catch (err) {
        console.error("Failed to fetch accounts:", err);
      }
    };
    fetchAccounts();
  }, []);

  const calculateTotals = () => {
    const totalDebit = lineItems.reduce((sum, item) => sum + (parseFloat(item.debit_amount) || 0), 0);
    const totalCredit = lineItems.reduce((sum, item) => sum + (parseFloat(item.credit_amount) || 0), 0);
    const difference = totalDebit - totalCredit;
    return { totalDebit, totalCredit, difference, isBalanced: Math.abs(difference) < 0.01 };
  };

  const addLineItem = () => {
    setLineItems([
      ...lineItems,
      {
        id: Date.now().toString(),
        account_no: "",
        debit_amount: "",
        credit_amount: "",
        tax_code: 0,
        description: "",
      },
    ]);
  };

  const removeLineItem = (id: string) => {
    if (lineItems.length > 2) {
      setLineItems(lineItems.filter((item) => item.id !== id));
    }
  };

  const updateLineItem = (id: string, field: keyof LineItemForm, value: string | number) => {
    setLineItems(
      lineItems.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    const { isBalanced, totalDebit } = calculateTotals();

    if (!isBalanced) {
      setError("Verifikatet är inte balanserat. Debet och kredit måste vara lika.");
      return;
    }

    if (totalDebit === 0) {
      setError("Verifikatet måste ha minst en transaktion.");
      return;
    }

    const filledLineItems = lineItems.filter(
      (item) => item.account_no && (parseFloat(item.debit_amount) > 0 || parseFloat(item.credit_amount) > 0)
    );

    if (filledLineItems.length < 2) {
      setError("Verifikatet måste ha minst två rader.");
      return;
    }

    setLoading(true);

    try {
      // Verify user is logged in
      if (!user) {
        setError("Du måste vara inloggad för att skapa ett verifikat.");
        setLoading(false);
        return;
      }

      // Create voucher
      // Extract period from date string (YYYY-MM-DD)
      const [year, month] = formData.date.split('-');
      const period = `${year}-${month}`;

      const voucher = await vouchersApi.create({
        date: formData.date, // Send as simple date string (YYYY-MM-DD)
        description: formData.description,
        reference: formData.reference,
        total_amount: totalDebit,
        period,
        created_by: user.user_id,
      });

      // Create line items
      for (const item of filledLineItems) {
        await lineItemsApi.create({
          voucher_id: voucher.voucher_id,
          account_no: parseInt(item.account_no),
          debit_amount: parseFloat(item.debit_amount) || 0,
          credit_amount: parseFloat(item.credit_amount) || 0,
          tax_code: item.tax_code,
        });
      }

      router.push("/vouchers");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create voucher");
    } finally {
      setLoading(false);
    }
  };

  const { totalDebit, totalCredit, difference, isBalanced } = calculateTotals();

  return (
    <DashboardLayout>
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => router.back()}
            className="text-sm text-gray-600 hover:text-gray-900 mb-4 flex items-center gap-2"
          >
            ← Tillbaka
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Nytt verifikat</h1>
          <p className="text-gray-600 mt-2">Skapa ett nytt bokföringsverifikat</p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Voucher Header */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Verifikatuppgifter</h2>

            {error && (
              <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <label htmlFor="date" className="block text-sm font-medium text-gray-700 mb-2">
                  Datum *
                </label>
                <input
                  id="date"
                  type="date"
                  required
                  value={formData.date}
                  onChange={(e) => setFormData({ ...formData, date: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
                />
              </div>

              <div>
                <label htmlFor="reference" className="block text-sm font-medium text-gray-700 mb-2">
                  Referens
                </label>
                <input
                  id="reference"
                  type="text"
                  value={formData.reference}
                  onChange={(e) => setFormData({ ...formData, reference: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
                  placeholder="t.ex. Fakturanummer"
                />
              </div>

              <div>
                <label htmlFor="period" className="block text-sm font-medium text-gray-700 mb-2">
                  Period
                </label>
                <input
                  id="period"
                  type="text"
                  disabled
                  value={`${new Date(formData.date).getFullYear()}-${String(new Date(formData.date).getMonth() + 1).padStart(2, "0")}`}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-100 text-gray-600"
                />
              </div>
            </div>

            <div className="mt-4">
              <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
                Beskrivning *
              </label>
              <textarea
                id="description"
                required
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                rows={2}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
                placeholder="Beskriv transaktionen..."
              />
            </div>
          </div>

          {/* Line Items */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Konteringsrader</h2>
              <button
                type="button"
                onClick={addLineItem}
                className="px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
              >
                + Lägg till rad
              </button>
            </div>

            <div className="space-y-4">
              {lineItems.map((item, index) => (
                <div key={item.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-semibold text-gray-500">Rad {index + 1}</span>
                    {lineItems.length > 2 && (
                      <button type="button" onClick={() => removeLineItem(item.id)}
                        className="text-red-500 hover:text-red-700 text-sm font-medium">
                        Ta bort
                      </button>
                    )}
                  </div>

                  <div className="space-y-3">
                    <div>
                      <label className="block text-xs font-medium text-gray-600 mb-1">Konto</label>
                      <AccountSearch
                        accounts={accounts}
                        value={item.account_no}
                        onChange={(accountNo) => updateLineItem(item.id, "account_no", accountNo)}
                        placeholder="Sök konto..."
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-medium text-gray-600 mb-1">Debet</label>
                      <input type="number" step="0.01" value={item.debit_amount} placeholder="0.00"
                        onChange={(e) => updateLineItem(item.id, "debit_amount", e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm text-gray-900" />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-600 mb-1">Kredit</label>
                      <input type="number" step="0.01" value={item.credit_amount} placeholder="0.00"
                        onChange={(e) => updateLineItem(item.id, "credit_amount", e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm text-gray-900" />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-600 mb-1">Moms</label>
                      <select value={item.tax_code}
                        onChange={(e) => updateLineItem(item.id, "tax_code", parseInt(e.target.value))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm text-gray-900">
                        <option value={0}>0%</option>
                        <option value={6}>6%</option>
                        <option value={12}>12%</option>
                        <option value={25}>25%</option>
                      </select>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Totals */}
            <div className="mt-4 border-t pt-4">
              <div className="flex justify-between text-sm font-semibold text-gray-900">
                <span>Summa:</span>
                <span>
                  {totalDebit.toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr (D) / {totalCredit.toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr (K)
                </span>
              </div>
              <div className={`flex justify-between text-sm font-semibold mt-2 ${isBalanced ? "text-green-600" : "text-red-600"}`}>
                <span>Differens:</span>
                <span>
                  {Math.abs(difference).toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr
                  {isBalanced ? " ✓ Balanserat" : " ✗ Ej balanserat"}
                </span>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-4">
            <button
              type="button"
              onClick={() => router.back()}
              className="flex-1 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium"
            >
              Avbryt
            </button>
            <button
              type="submit"
              disabled={loading || !isBalanced}
              className="flex-1 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Sparar..." : "Skapa verifikat"}
            </button>
          </div>
        </form>

        {/* Help Text */}
        <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="text-sm font-semibold text-blue-900 mb-2">💡 Tips för dubbel bokföring</h3>
          <ul className="text-sm text-blue-700 space-y-1">
            <li>• Varje verifikat måste vara balanserat (debet = kredit)</li>
            <li>• Tillgångar och kostnader bokförs på debet</li>
            <li>• Skulder, eget kapital och intäkter bokförs på kredit</li>
            <li>• Ett verifikat måste ha minst två rader</li>
          </ul>
        </div>
      </div>
    </DashboardLayout>
  );
}
