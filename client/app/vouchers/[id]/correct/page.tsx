"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { vouchersApi } from "@/lib/api/vouchers";
import { lineItemsApi } from "@/lib/api/lineitems";
import { accountsApi } from "@/lib/api/accounts";
import { Account, LineItem } from "@/types";
import { useAuth } from "@/lib/contexts/AuthContext";
import AccountSearch from "@/components/ui/AccountSearch";

interface LineItemForm {
  id: string;
  account_no: string;
  debit_amount: string;
  credit_amount: string;
  tax_code: number;
}

export default function CorrectVoucherPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const voucherId = Number(params.id);

  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [originalVoucherNumber, setOriginalVoucherNumber] = useState(0);

  const [formData, setFormData] = useState({
    date: new Date().toISOString().split("T")[0],
    description: "",
    reference: "",
  });

  const [lineItems, setLineItems] = useState<LineItemForm[]>([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [voucherData, lineItemsData, accountsData] = await Promise.all([
          vouchersApi.getById(voucherId),
          lineItemsApi.getByVoucher(voucherId),
          accountsApi.getAll(),
        ]);

        setAccounts(accountsData || []);
        setOriginalVoucherNumber(voucherData.voucher_number);

        // Pre-fill form with original data
        setFormData({
          date: voucherData.date.split("T")[0],
          description: voucherData.description,
          reference: voucherData.reference || "",
        });

        // Convert line items to form format
        const formLineItems = (lineItemsData || []).map((item: LineItem) => ({
          id: item.line_id.toString(),
          account_no: item.account_no.toString(),
          debit_amount: item.debit_amount.toString(),
          credit_amount: item.credit_amount.toString(),
          tax_code: item.tax_code,
        }));

        setLineItems(formLineItems);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load voucher");
      } finally {
        setLoading(false);
      }
    };

    if (voucherId && user) {
      fetchData();
    }
  }, [voucherId, user]);

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

    if (!user) {
      setError("Du måste vara inloggad.");
      return;
    }

    setSubmitting(true);

    try {
      // Extract period from date
      const [year, month] = formData.date.split('-');
      const period = `${year}-${month}`;

      // Send correction data to backend
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/vouchers/${voucherId}/correct-with-changes`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          user_id: user.user_id,
          new_voucher: {
            date: formData.date,
            description: formData.description,
            reference: formData.reference,
            total_amount: totalDebit,
            period,
            created_by: user.user_id,
          },
          new_line_items: filledLineItems.map(item => ({
            account_no: parseInt(item.account_no),
            debit_amount: parseFloat(item.debit_amount) || 0,
            credit_amount: parseFloat(item.credit_amount) || 0,
            tax_code: item.tax_code,
          })),
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to create correction");
      }

      const correctionVoucher = await response.json();
      router.push(`/vouchers/${correctionVoucher.voucher_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create correction");
    } finally {
      setSubmitting(false);
    }
  };

  const { totalDebit, totalCredit, difference, isBalanced } = calculateTotals();

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
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-gray-900">
            Skapa rättelse av verifikat #{originalVoucherNumber}
          </h1>
          <p className="text-gray-600 mt-2">
            Redigera värdena nedan. Ett rättelseverifikat kommer att skapas som reverserar det gamla och lägger till det nya.
          </p>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Voucher Info */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">Verifikatuppgifter</h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label htmlFor="date" className="block text-sm font-medium text-gray-700 mb-2">
                  Datum *
                </label>
                <input
                  type="date"
                  id="date"
                  value={formData.date}
                  onChange={(e) => setFormData({ ...formData, date: e.target.value })}
                  required
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
                />
              </div>

              <div>
                <label htmlFor="reference" className="block text-sm font-medium text-gray-700 mb-2">
                  Referens
                </label>
                <input
                  type="text"
                  id="reference"
                  value={formData.reference}
                  onChange={(e) => setFormData({ ...formData, reference: e.target.value })}
                  placeholder="t.ex. Fakturanummer"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
                />
              </div>

              <div>
                <label htmlFor="period" className="block text-sm font-medium text-gray-700 mb-2">
                  Period
                </label>
                <input
                  type="text"
                  id="period"
                  value={formData.date ? `${formData.date.split('-')[0]}-${formData.date.split('-')[1]}` : ''}
                  disabled
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-600"
                />
              </div>
            </div>

            <div className="mt-4">
              <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
                Beskrivning *
              </label>
              <textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                required
                rows={3}
                placeholder="Beskriv transaktionen..."
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
              />
            </div>
          </div>

          {/* Line Items */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold text-gray-900">Konteringsrader</h2>
              <button
                type="button"
                onClick={addLineItem}
                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm font-medium"
              >
                + Lägg till rad
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Konto</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Kontonamn</th>
                    <th className="text-right py-3 px-4 text-sm font-semibold text-gray-700">Debet</th>
                    <th className="text-right py-3 px-4 text-sm font-semibold text-gray-700">Kredit</th>
                    <th className="text-center py-3 px-4 text-sm font-semibold text-gray-700">Moms</th>
                    <th className="text-center py-3 px-4 text-sm font-semibold text-gray-700"></th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {lineItems.map((item, index) => {
                    const account = accounts.find((a) => a.account_no === parseInt(item.account_no));
                    return (
                      <tr key={item.id} className="hover:bg-gray-50">
                        <td className="py-3 px-4">
                          <AccountSearch
                            accounts={accounts}
                            value={item.account_no}
                            onChange={(accountNo) => updateLineItem(item.id, "account_no", accountNo)}
                          />
                        </td>
                        <td className="py-3 px-4 text-sm text-gray-600">
                          {account ? account.account_name : "-"}
                        </td>
                        <td className="py-3 px-4">
                          <input
                            type="number"
                            step="0.01"
                            value={item.debit_amount}
                            onChange={(e) => updateLineItem(item.id, "debit_amount", e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-right text-gray-900"
                            placeholder="0.00"
                          />
                        </td>
                        <td className="py-3 px-4">
                          <input
                            type="number"
                            step="0.01"
                            value={item.credit_amount}
                            onChange={(e) => updateLineItem(item.id, "credit_amount", e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-right text-gray-900"
                            placeholder="0.00"
                          />
                        </td>
                        <td className="py-3 px-4">
                          <select
                            value={item.tax_code}
                            onChange={(e) => updateLineItem(item.id, "tax_code", parseInt(e.target.value))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-gray-900"
                          >
                            <option value={0}>0%</option>
                            <option value={6}>6%</option>
                            <option value={12}>12%</option>
                            <option value={25}>25%</option>
                          </select>
                        </td>
                        <td className="py-3 px-4 text-center">
                          {lineItems.length > 2 && (
                            <button
                              type="button"
                              onClick={() => removeLineItem(item.id)}
                              className="text-red-600 hover:text-red-700 font-medium text-xl"
                            >
                              ×
                            </button>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
                <tfoot className="bg-gray-50 font-semibold">
                  <tr>
                    <td colSpan={2} className="py-3 px-4 text-right">Summa:</td>
                    <td className="py-3 px-4 text-right">{totalDebit.toFixed(2)} kr</td>
                    <td className="py-3 px-4 text-right">{totalCredit.toFixed(2)} kr</td>
                    <td colSpan={2}></td>
                  </tr>
                </tfoot>
              </table>
            </div>

            {/* Balance Status */}
            <div className="mt-4 flex items-center justify-between p-4 rounded-lg bg-gray-50">
              <div>
                <span className="text-sm text-gray-600">Differens:</span>
                <span className={`ml-2 font-semibold ${Math.abs(difference) < 0.01 ? 'text-green-600' : 'text-red-600'}`}>
                  {difference.toFixed(2)} kr
                </span>
              </div>
              <div className="flex items-center gap-2">
                {isBalanced ? (
                  <>
                    <span className="text-green-600 font-semibold">✓ Balanserat</span>
                  </>
                ) : (
                  <span className="text-red-600 font-semibold">⚠ Inte balanserat</span>
                )}
              </div>
            </div>
          </div>

          {/* Submit Buttons */}
          <div className="flex gap-4">
            <button
              type="button"
              onClick={() => router.back()}
              disabled={submitting}
              className="px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium"
            >
              Avbryt
            </button>
            <button
              type="submit"
              disabled={submitting || !isBalanced}
              className="flex-1 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? "Skapar rättelse..." : "Skapa rättelse"}
            </button>
          </div>
        </form>
      </div>
    </DashboardLayout>
  );
}
