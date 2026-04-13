"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { invoicesApi } from "@/lib/api/invoices";
import { customersApi } from "@/lib/api/customers";
import { Customer } from "@/types";
import { formatSEK } from "@/lib/money";
import Link from "next/link";

interface LineForm {
  id: string;
  description: string;
  quantity: string;
  unit: string;
  unit_price: string;
  discount_percent: string;
  vat_rate: number;
  account_no: string;
}

const emptyLine = (): LineForm => ({
  id: String(Date.now() + Math.random()),
  description: "",
  quantity: "1",
  unit: "st",
  unit_price: "",
  discount_percent: "0",
  vat_rate: 25,
  account_no: "",
});

export default function NewInvoicePage() {
  const router = useRouter();
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const [form, setForm] = useState({
    customer_id: "",
    invoice_date: new Date().toISOString().split("T")[0],
    payment_terms_days: "30",
    notes: "",
    your_reference: "",
    our_reference: "",
  });

  const [lines, setLines] = useState<LineForm[]>([emptyLine()]);

  useEffect(() => {
    customersApi.getAll().then((data) => setCustomers(data || []));
  }, []);

  const addLine = () => setLines([...lines, emptyLine()]);

  const removeLine = (id: string) => {
    if (lines.length <= 1) return;
    setLines(lines.filter((l) => l.id !== id));
  };

  const updateLine = (id: string, field: string, value: string | number) => {
    setLines(lines.map((l) => (l.id === id ? { ...l, [field]: value } : l)));
  };

  // Calculate totals client-side for preview
  const calcLine = (l: LineForm) => {
    const qty = parseFloat(l.quantity) || 0;
    const price = parseFloat(l.unit_price) || 0;
    const discount = parseFloat(l.discount_percent) || 0;
    const lineTotal = qty * price * (1 - discount / 100);
    const vatAmount = lineTotal * (l.vat_rate / 100);
    return { lineTotal, vatAmount };
  };

  const subtotal = lines.reduce((sum, l) => sum + calcLine(l).lineTotal, 0);
  const vatTotal = lines.reduce((sum, l) => sum + calcLine(l).vatAmount, 0);
  const rawTotal = subtotal + vatTotal;
  const roundedTotal = Math.round(rawTotal);
  const rounding = roundedTotal - rawTotal;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.customer_id) {
      setError("Välj en kund");
      return;
    }
    if (lines.every((l) => !l.description || !l.unit_price)) {
      setError("Lägg till minst en rad med beskrivning och pris");
      return;
    }

    setSaving(true);
    setError("");
    try {
      const invoice = await invoicesApi.create({
        customer_id: Number(form.customer_id),
        invoice_date: form.invoice_date,
        payment_terms_days: Number(form.payment_terms_days) || 30,
        notes: form.notes,
        your_reference: form.your_reference,
        our_reference: form.our_reference,
        lines: lines
          .filter((l) => l.description && l.unit_price)
          .map((l) => ({
            description: l.description,
            quantity: l.quantity || "1",
            unit: l.unit || "st",
            unit_price: l.unit_price,
            discount_percent: l.discount_percent || "0",
            vat_rate: l.vat_rate,
            account_no: l.account_no ? Number(l.account_no) : undefined,
          })),
      });
      router.push(`/invoices/${invoice.invoice_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte skapa faktura");
    } finally {
      setSaving(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="mb-8">
        <Link href="/invoices" className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block">
          &larr; Tillbaka till fakturor
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">Ny faktura</h1>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">{error}</div>
      )}

      <form onSubmit={handleSubmit}>
        {/* Header */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Fakturainformation</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Kund *</label>
              <select
                value={form.customer_id}
                onChange={(e) => setForm({ ...form, customer_id: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
              >
                <option value="">Välj kund...</option>
                {customers.map((c) => (
                  <option key={c.customer_id} value={c.customer_id}>
                    {c.name} {c.org_number ? `(${c.org_number})` : ""}
                  </option>
                ))}
              </select>
              {customers.length === 0 && (
                <p className="text-sm text-gray-500 mt-1">
                  <Link href="/customers/new" className="text-blue-600 hover:underline">Skapa en kund</Link> först
                </p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Fakturadatum</label>
              <input type="date" value={form.invoice_date}
                onChange={(e) => setForm({ ...form, invoice_date: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Betalningsvillkor (dagar)</label>
              <input type="number" value={form.payment_terms_days}
                onChange={(e) => setForm({ ...form, payment_terms_days: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Er referens</label>
              <input value={form.your_reference}
                onChange={(e) => setForm({ ...form, your_reference: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Vår referens</label>
              <input value={form.our_reference}
                onChange={(e) => setForm({ ...form, our_reference: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
          </div>
        </div>

        {/* Line items */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Fakturarader</h2>
            <button type="button" onClick={addLine}
              className="px-3 py-1.5 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm font-medium">
              + Lägg till rad
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left py-2 px-3 text-xs font-semibold text-gray-700 w-1/3">Beskrivning</th>
                  <th className="text-right py-2 px-3 text-xs font-semibold text-gray-700 w-16">Antal</th>
                  <th className="text-left py-2 px-3 text-xs font-semibold text-gray-700 w-16">Enhet</th>
                  <th className="text-right py-2 px-3 text-xs font-semibold text-gray-700 w-24">Á-pris</th>
                  <th className="text-right py-2 px-3 text-xs font-semibold text-gray-700 w-20">Rabatt %</th>
                  <th className="text-center py-2 px-3 text-xs font-semibold text-gray-700 w-20">Moms</th>
                  <th className="text-right py-2 px-3 text-xs font-semibold text-gray-700 w-24">Summa</th>
                  <th className="w-10"></th>
                </tr>
              </thead>
              <tbody>
                {lines.map((line) => {
                  const { lineTotal, vatAmount } = calcLine(line);
                  return (
                    <tr key={line.id} className="border-b border-gray-100">
                      <td className="py-2 px-3">
                        <input value={line.description} placeholder="Beskrivning..."
                          onChange={(e) => updateLine(line.id, "description", e.target.value)}
                          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm text-gray-900" />
                      </td>
                      <td className="py-2 px-3">
                        <input type="number" step="any" value={line.quantity}
                          onChange={(e) => updateLine(line.id, "quantity", e.target.value)}
                          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm text-gray-900 text-right" />
                      </td>
                      <td className="py-2 px-3">
                        <input value={line.unit}
                          onChange={(e) => updateLine(line.id, "unit", e.target.value)}
                          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm text-gray-900" />
                      </td>
                      <td className="py-2 px-3">
                        <input type="number" step="0.01" value={line.unit_price} placeholder="0.00"
                          onChange={(e) => updateLine(line.id, "unit_price", e.target.value)}
                          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm text-gray-900 text-right" />
                      </td>
                      <td className="py-2 px-3">
                        <input type="number" step="0.01" value={line.discount_percent}
                          onChange={(e) => updateLine(line.id, "discount_percent", e.target.value)}
                          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm text-gray-900 text-right" />
                      </td>
                      <td className="py-2 px-3">
                        <select value={line.vat_rate}
                          onChange={(e) => updateLine(line.id, "vat_rate", Number(e.target.value))}
                          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm text-gray-900">
                          <option value={25}>25%</option>
                          <option value={12}>12%</option>
                          <option value={6}>6%</option>
                          <option value={0}>0%</option>
                        </select>
                      </td>
                      <td className="py-2 px-3 text-right text-sm text-gray-900 font-medium">
                        {formatSEK(lineTotal)}
                      </td>
                      <td className="py-2 px-1">
                        <button type="button" onClick={() => removeLine(line.id)}
                          className="text-red-500 hover:text-red-700 text-lg font-bold px-2">
                          &times;
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Totals */}
          <div className="mt-4 border-t pt-4 flex justify-end">
            <div className="w-72 space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">Netto:</span>
                <span className="font-medium text-gray-900">{formatSEK(subtotal)} kr</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Moms:</span>
                <span className="font-medium text-gray-900">{formatSEK(vatTotal)} kr</span>
              </div>
              {Math.abs(rounding) > 0.001 && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Öresavrundning:</span>
                  <span className="font-medium text-gray-900">{formatSEK(rounding)} kr</span>
                </div>
              )}
              <div className="flex justify-between border-t pt-2 text-base">
                <span className="font-semibold text-gray-900">Att betala:</span>
                <span className="font-bold text-gray-900">{formatSEK(roundedTotal)} kr</span>
              </div>
            </div>
          </div>
        </div>

        {/* Notes */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-1">Meddelande till kund</label>
          <textarea value={form.notes} rows={2}
            onChange={(e) => setForm({ ...form, notes: e.target.value })}
            placeholder="Valfritt meddelande som visas på fakturan..."
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
        </div>

        {/* Actions */}
        <div className="flex gap-4">
          <Link href="/invoices"
            className="flex-1 text-center px-4 py-3 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors font-medium">
            Avbryt
          </Link>
          <button type="submit" disabled={saving}
            className="flex-1 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50">
            {saving ? "Sparar..." : "Skapa faktura (utkast)"}
          </button>
        </div>
      </form>
    </DashboardLayout>
  );
}
