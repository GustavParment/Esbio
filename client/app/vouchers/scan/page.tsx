"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { receiptsApi } from "@/lib/api/receipts";
import { vouchersApi } from "@/lib/api/vouchers";
import { accountsApi } from "@/lib/api/accounts";
import { lineItemsApi } from "@/lib/api/lineitems";
import { Account, ReceiptScanResult } from "@/types";
import { formatSEK, parseMoney } from "@/lib/money";
import { useAuth } from "@/lib/contexts/AuthContext";
import AccountSearch from "@/components/ui/AccountSearch";

interface LineItemForm {
  id: string;
  account_no: string;
  debit_amount: string;
  credit_amount: string;
  tax_code: number;
}

export default function ScanReceiptPage() {
  const router = useRouter();
  const { user } = useAuth();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [scanning, setScanning] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [preview, setPreview] = useState<string | null>(null);
  const [previewKind, setPreviewKind] = useState<"image" | "pdf" | null>(null);
  const [previewName, setPreviewName] = useState<string>("");
  const [scanResult, setScanResult] = useState<ReceiptScanResult | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);

  const [formData, setFormData] = useState({
    date: new Date().toISOString().split("T")[0],
    description: "",
    reference: "",
  });

  const [lineItems, setLineItems] = useState<LineItemForm[]>([
    { id: "1", account_no: "", debit_amount: "", credit_amount: "", tax_code: 0 },
    { id: "2", account_no: "", debit_amount: "", credit_amount: "", tax_code: 0 },
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

  const handleFileSelect = async (file: File) => {
    setError("");
    setScanResult(null);

    const isPdf = file.type === "application/pdf";
    const isImage = file.type.startsWith("image/");
    if (!isPdf && !isImage) {
      setError("Ogiltigt filformat. Ladda upp en bild (JPEG, PNG, WebP, HEIC) eller PDF.");
      return;
    }

    setPreviewKind(isPdf ? "pdf" : "image");
    setPreviewName(file.name);

    if (isImage) {
      const reader = new FileReader();
      reader.onload = (e) => setPreview(e.target?.result as string);
      reader.readAsDataURL(file);
    } else {
      setPreview(null);
    }

    // Scan
    setScanning(true);
    try {
      const result = await receiptsApi.scan(file);
      setScanResult(result);

      // Pre-fill form
      setFormData({
        date: result.date || new Date().toISOString().split("T")[0],
        description: result.vendor ? `${result.vendor} - ${result.description}` : result.description,
        reference: "",
      });

      // Pre-fill line items from AI suggestion
      if (result.suggested_lines && result.suggested_lines.length > 0) {
        setLineItems(
          result.suggested_lines.map((line, i) => ({
            id: String(i + 1),
            account_no: String(line.account_no),
            debit_amount: parseMoney(line.debit_amount) > 0 ? String(line.debit_amount) : "",
            credit_amount: parseMoney(line.credit_amount) > 0 ? String(line.credit_amount) : "",
            tax_code: line.tax_code,
          }))
        );
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte läsa kvittot");
    } finally {
      setScanning(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file && (file.type.startsWith("image/") || file.type === "application/pdf")) {
      handleFileSelect(file);
    }
  };

  const calculateTotals = () => {
    const totalDebit = lineItems.reduce((sum, item) => sum + (parseFloat(item.debit_amount) || 0), 0);
    const totalCredit = lineItems.reduce((sum, item) => sum + (parseFloat(item.credit_amount) || 0), 0);
    const difference = totalDebit - totalCredit;
    return { totalDebit, totalCredit, difference, isBalanced: Math.abs(difference) < 0.01 };
  };

  const addLineItem = () => {
    setLineItems([
      ...lineItems,
      { id: Date.now().toString(), account_no: "", debit_amount: "", credit_amount: "", tax_code: 0 },
    ]);
  };

  const removeLineItem = (id: string) => {
    if (lineItems.length > 2) {
      setLineItems(lineItems.filter((item) => item.id !== id));
    }
  };

  const updateLineItem = (id: string, field: keyof LineItemForm, value: string | number) => {
    setLineItems(lineItems.map((item) => (item.id === id ? { ...item, [field]: value } : item)));
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

    setSaving(true);
    try {
      const [year, month] = formData.date.split("-");
      const period = `${year}-${month}`;

      const voucher = await vouchersApi.create({
        date: formData.date,
        description: formData.description,
        reference: formData.reference,
        total_amount: totalDebit,
        period,
        created_by: user.user_id,
      });

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
      setError(err instanceof Error ? err.message : "Kunde inte skapa verifikatet");
    } finally {
      setSaving(false);
    }
  };

  const { totalDebit, totalCredit, difference, isBalanced } = calculateTotals();

  return (
    <DashboardLayout>
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <button onClick={() => router.back()} className="text-sm text-gray-600 hover:text-gray-900 mb-4 flex items-center gap-2">
            ← Tillbaka
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Skanna kvitto eller PDF</h1>
          <p className="text-gray-600 mt-2">Ladda upp ett kvitto, faktura eller PDF — AI:n läser av och föreslår kontering</p>
        </div>

        {/* Upload area */}
        {!scanResult && (
          <div
            onDrop={handleDrop}
            onDragOver={(e) => e.preventDefault()}
            onClick={() => fileInputRef.current?.click()}
            className={`bg-white rounded-xl shadow-sm border-2 border-dashed p-12 text-center cursor-pointer transition-colors ${
              scanning ? "border-blue-400 bg-blue-50" : "border-gray-300 hover:border-blue-400 hover:bg-gray-50"
            }`}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*,application/pdf"
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) handleFileSelect(file);
              }}
            />

            {scanning ? (
              <div>
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto mb-4" />
                <p className="text-lg font-medium text-blue-700">AI:n analyserar dokumentet...</p>
                <p className="text-sm text-blue-500 mt-1">Detta kan ta några sekunder</p>
              </div>
            ) : (
              <div>
                <div className="text-5xl mb-4">📄</div>
                <p className="text-lg font-medium text-gray-900">Dra och släpp ett kvitto eller PDF här</p>
                <p className="text-sm text-gray-500 mt-2">eller klicka för att välja fil (JPEG, PNG, WebP, HEIC, PDF)</p>
                <p className="text-xs text-gray-400 mt-1">Max 20 MB</p>
              </div>
            )}
          </div>
        )}

        {error && !scanResult && (
          <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* Result: Preview + Form */}
        {scanResult && (
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* AI result summary */}
            <div className="bg-green-50 border border-green-200 rounded-xl p-6">
              <div className="flex items-start gap-4">
                {previewKind === "image" && preview && (
                  <img src={preview} alt="Kvitto" className="w-24 h-24 object-cover rounded-lg border border-green-200 flex-shrink-0" />
                )}
                {previewKind === "pdf" && (
                  <div className="w-24 h-24 rounded-lg border border-green-200 flex-shrink-0 bg-white flex flex-col items-center justify-center text-xs text-gray-600 p-2">
                    <div className="text-3xl mb-1">📄</div>
                    <div className="truncate w-full text-center" title={previewName}>PDF</div>
                  </div>
                )}
                <div className="flex-1">
                  <h3 className="font-semibold text-green-900 text-lg">AI-analys klar</h3>
                  <div className="mt-2 grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
                    <div>
                      <span className="text-green-700">Leverantör:</span>
                      <p className="font-medium text-green-900">{scanResult.vendor}</p>
                    </div>
                    <div>
                      <span className="text-green-700">Belopp:</span>
                      <p className="font-medium text-green-900">
                        {formatSEK(scanResult.total_amount)} {scanResult.currency}
                      </p>
                    </div>
                    <div>
                      <span className="text-green-700">Moms ({scanResult.vat_rate}%):</span>
                      <p className="font-medium text-green-900">
                        {formatSEK(scanResult.vat_amount)} {scanResult.currency}
                      </p>
                    </div>
                    <div>
                      <span className="text-green-700">Datum:</span>
                      <p className="font-medium text-green-900">{scanResult.date}</p>
                    </div>
                  </div>
                </div>
              </div>
              <button
                type="button"
                onClick={() => {
                  setScanResult(null);
                  setPreview(null);
                  setPreviewKind(null);
                  setPreviewName("");
                  setFormData({ date: new Date().toISOString().split("T")[0], description: "", reference: "" });
                  setLineItems([
                    { id: "1", account_no: "", debit_amount: "", credit_amount: "", tax_code: 0 },
                    { id: "2", account_no: "", debit_amount: "", credit_amount: "", tax_code: 0 },
                  ]);
                }}
                className="mt-3 text-sm text-green-700 hover:text-green-900 underline"
              >
                Skanna nytt kvitto
              </button>
            </div>

            {/* Voucher form */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Verifikatuppgifter</h2>

              {error && (
                <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
                  <p className="text-sm text-red-600">{error}</p>
                </div>
              )}

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <label htmlFor="date" className="block text-sm font-medium text-gray-700 mb-2">Datum *</label>
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
                  <label htmlFor="reference" className="block text-sm font-medium text-gray-700 mb-2">Referens</label>
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
                  <label htmlFor="period" className="block text-sm font-medium text-gray-700 mb-2">Period</label>
                  <input
                    id="period"
                    type="text"
                    disabled
                    value={formData.date ? `${formData.date.split("-")[0]}-${formData.date.split("-")[1]}` : ""}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-100 text-gray-600"
                  />
                </div>
              </div>

              <div className="mt-4">
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">Beskrivning *</label>
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
                <h2 className="text-lg font-semibold text-gray-900">Konteringsrader (AI-förslag)</h2>
                <button type="button" onClick={addLineItem} className="px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors">
                  + Lägg till rad
                </button>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-gray-200">
                      <th className="text-left py-3 px-2 text-sm font-semibold text-gray-700">Konto</th>
                      <th className="text-right py-3 px-2 text-sm font-semibold text-gray-700">Debet</th>
                      <th className="text-right py-3 px-2 text-sm font-semibold text-gray-700">Kredit</th>
                      <th className="text-left py-3 px-2 text-sm font-semibold text-gray-700">Moms</th>
                      <th className="py-3 px-2"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {lineItems.map((item) => (
                      <tr key={item.id} className="border-b border-gray-100">
                        <td className="py-3 px-2 min-w-[280px]">
                          <AccountSearch
                            accounts={accounts}
                            value={item.account_no}
                            onChange={(accountNo) => updateLineItem(item.id, "account_no", accountNo)}
                            placeholder="Sök konto..."
                          />
                        </td>
                        <td className="py-3 px-2">
                          <input
                            type="number"
                            step="0.01"
                            value={item.debit_amount}
                            onChange={(e) => updateLineItem(item.id, "debit_amount", e.target.value)}
                            className="w-full px-2 py-2 border border-gray-300 rounded text-sm text-right text-black focus:ring-2 focus:ring-blue-500"
                            placeholder="0.00"
                          />
                        </td>
                        <td className="py-3 px-2">
                          <input
                            type="number"
                            step="0.01"
                            value={item.credit_amount}
                            onChange={(e) => updateLineItem(item.id, "credit_amount", e.target.value)}
                            className="w-full px-2 py-2 border border-gray-300 rounded text-sm text-right text-black focus:ring-2 focus:ring-blue-500"
                            placeholder="0.00"
                          />
                        </td>
                        <td className="py-3 px-2">
                          <select
                            value={item.tax_code}
                            onChange={(e) => updateLineItem(item.id, "tax_code", parseInt(e.target.value))}
                            className="w-full px-2 py-2 border border-gray-300 rounded text-sm text-black focus:ring-2 focus:ring-blue-500"
                          >
                            <option value={0}>0%</option>
                            <option value={6}>6%</option>
                            <option value={12}>12%</option>
                            <option value={25}>25%</option>
                          </select>
                        </td>
                        <td className="py-3 px-2 text-center">
                          {lineItems.length > 2 && (
                            <button type="button" onClick={() => removeLineItem(item.id)} className="text-red-600 hover:text-red-700 text-sm">
                              ✕
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot>
                    <tr className="border-t-2 border-gray-300 font-semibold">
                      <td className="py-3 px-2 text-sm text-gray-900">Summa:</td>
                      <td className="py-3 px-2 text-right text-sm text-gray-900">
                        {totalDebit.toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr
                      </td>
                      <td className="py-3 px-2 text-right text-sm text-gray-900">
                        {totalCredit.toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr
                      </td>
                      <td colSpan={2}></td>
                    </tr>
                    <tr>
                      <td className="py-2 px-2 text-sm font-medium text-gray-700">Differens:</td>
                      <td colSpan={2} className={`py-2 px-2 text-right text-sm font-semibold ${isBalanced ? "text-green-600" : "text-red-600"}`}>
                        {Math.abs(difference).toLocaleString("sv-SE", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} kr
                        {isBalanced ? " ✓ Balanserat" : " ✗ Ej balanserat"}
                      </td>
                      <td colSpan={2}></td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>

            {/* Actions */}
            <div className="flex gap-4">
              <button type="button" onClick={() => router.back()} className="flex-1 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium">
                Avbryt
              </button>
              <button
                type="submit"
                disabled={saving || !isBalanced}
                className="flex-1 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {saving ? "Sparar..." : "Skapa verifikat"}
              </button>
            </div>
          </form>
        )}
      </div>
    </DashboardLayout>
  );
}
