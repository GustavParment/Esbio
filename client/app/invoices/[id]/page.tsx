"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { invoicesApi } from "@/lib/api/invoices";
import { Invoice, InvoiceStatus } from "@/types";
import { formatSEK, parseMoney } from "@/lib/money";
import Link from "next/link";

const STATUS_LABELS: Record<InvoiceStatus, string> = {
  draft: "Utkast", sent: "Skickad", paid: "Betald", overdue: "Förfallen", cancelled: "Makulerad",
};
const STATUS_COLORS: Record<InvoiceStatus, string> = {
  draft: "bg-gray-100 text-gray-700", sent: "bg-blue-100 text-blue-700",
  paid: "bg-green-100 text-green-700", overdue: "bg-red-100 text-red-700",
  cancelled: "bg-gray-100 text-gray-400",
};

export default function InvoiceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const invoiceId = Number(params.id);

  const [invoice, setInvoice] = useState<Invoice | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState("");
  const [paymentDate, setPaymentDate] = useState(new Date().toISOString().split("T")[0]);
  const [emailSending, setEmailSending] = useState(false);
  const [emailSent, setEmailSent] = useState(false);

  useEffect(() => {
    const fetchInvoice = async () => {
      try {
        const data = await invoicesApi.getById(invoiceId);
        setInvoice(data);
      } catch (err) {
        setError("Faktura hittades inte");
      } finally {
        setLoading(false);
      }
    };
    if (!isNaN(invoiceId)) fetchInvoice();
  }, [invoiceId]);

  const handleFinalize = async () => {
    if (!confirm("Skicka fakturan? Detta skapar ett bokföringsverifikat och kan inte ångras.")) return;
    setActionLoading(true);
    try {
      const updated = await invoicesApi.finalize(invoiceId);
      setInvoice(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte skicka faktura");
    } finally {
      setActionLoading(false);
    }
  };

  const handlePay = async () => {
    if (!paymentDate) { setError("Ange betalningsdatum"); return; }
    setActionLoading(true);
    try {
      const updated = await invoicesApi.markAsPaid(invoiceId, { payment_date: paymentDate });
      setInvoice(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte markera som betald");
    } finally {
      setActionLoading(false);
    }
  };

  const handleCancel = async () => {
    if (!confirm("Makulera fakturan? Ett rättelseverifikat skapas.")) return;
    setActionLoading(true);
    try {
      const updated = await invoicesApi.cancel(invoiceId);
      setInvoice(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte makulera");
    } finally {
      setActionLoading(false);
    }
  };

  const handleSendEmail = async () => {
    if (!invoice) return;
    const customerEmail = invoice.customer?.email;
    if (!customerEmail) { setError("Kunden saknar e-postadress"); return; }
    if (!confirm(`Skicka faktura ${invoice.invoice_number} till ${customerEmail}?`)) return;
    setEmailSending(true);
    setError("");
    try {
      await invoicesApi.sendEmail(invoiceId);
      setEmailSent(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte skicka e-post");
    } finally {
      setEmailSending(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Radera utkastet? Detta kan inte ångras.")) return;
    try {
      await invoicesApi.delete(invoiceId);
      router.push("/invoices");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte radera");
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  if (!invoice) {
    return (
      <DashboardLayout>
        <div className="text-center py-12">
          <p className="text-gray-500 mb-4">{error || "Faktura hittades inte"}</p>
          <Link href="/invoices" className="text-blue-600 hover:text-blue-700 font-medium">&larr; Tillbaka</Link>
        </div>
      </DashboardLayout>
    );
  }

  const status = invoice.status as InvoiceStatus;

  return (
    <DashboardLayout>
      <div className="mb-8">
        <Link href="/invoices" className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block">
          &larr; Tillbaka till fakturor
        </Link>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">
              Faktura #{invoice.invoice_number}
            </h1>
            <span className={`inline-flex mt-2 px-3 py-1 rounded-full text-sm font-medium ${STATUS_COLORS[status] || ""}`}>
              {STATUS_LABELS[status] || status}
            </span>
          </div>
          <div className="flex gap-3">
            {status !== "draft" && invoice.customer?.email && (
              <button
                onClick={handleSendEmail}
                disabled={emailSending || emailSent}
                className={`px-4 py-2.5 rounded-lg transition-colors font-medium text-sm ${
                  emailSent
                    ? "bg-green-100 text-green-700"
                    : "bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
                }`}
              >
                {emailSending ? "Skickar..." : emailSent ? "Skickad!" : "Skicka via e-post"}
              </button>
            )}
            <a
              href={`${process.env.NEXT_PUBLIC_API_URL || "https://api.esbio.se/api/v1"}/invoices/${invoice.invoice_id}/pdf`}
              target="_blank"
              rel="noopener noreferrer"
              className="px-4 py-2.5 bg-gray-700 text-white rounded-lg hover:bg-gray-800 transition-colors font-medium text-sm"
            >
              Ladda ner PDF
            </a>
          </div>
        </div>
      </div>

      {error && <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">{error}</div>}

      {/* Info grid */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          <div>
            <p className="text-sm text-gray-500 mb-1">Kund</p>
            <p className="font-medium text-gray-900">
              {invoice.customer ? (
                <Link href={`/customers/${invoice.customer.customer_id}`} className="text-blue-600 hover:underline">
                  {invoice.customer.name}
                </Link>
              ) : `Kund #${invoice.customer_id}`}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">Fakturadatum</p>
            <p className="font-medium text-gray-900">{new Date(invoice.invoice_date).toLocaleDateString("sv-SE")}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">Förfallodatum</p>
            <p className="font-medium text-gray-900">{new Date(invoice.due_date).toLocaleDateString("sv-SE")}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">OCR</p>
            <p className="font-mono font-medium text-gray-900">{invoice.ocr_number}</p>
          </div>
          {invoice.your_reference && (
            <div>
              <p className="text-sm text-gray-500 mb-1">Er referens</p>
              <p className="font-medium text-gray-900">{invoice.your_reference}</p>
            </div>
          )}
          {invoice.our_reference && (
            <div>
              <p className="text-sm text-gray-500 mb-1">Vår referens</p>
              <p className="font-medium text-gray-900">{invoice.our_reference}</p>
            </div>
          )}
          {invoice.bankgiro && (
            <div>
              <p className="text-sm text-gray-500 mb-1">Bankgiro</p>
              <p className="font-medium text-gray-900">{invoice.bankgiro}</p>
            </div>
          )}
          {invoice.plusgiro && (
            <div>
              <p className="text-sm text-gray-500 mb-1">Plusgiro</p>
              <p className="font-medium text-gray-900">{invoice.plusgiro}</p>
            </div>
          )}
        </div>
        {invoice.sales_voucher_id && (
          <div className="mt-4 pt-4 border-t text-sm text-gray-600">
            Bokföringsverifikat: <Link href={`/vouchers/${invoice.sales_voucher_id}`} className="text-blue-600 hover:underline font-medium">Visa verifikat</Link>
          </div>
        )}
        {invoice.payment_voucher_id && (
          <div className="mt-2 text-sm text-gray-600">
            Betalningsverifikat: <Link href={`/vouchers/${invoice.payment_voucher_id}`} className="text-blue-600 hover:underline font-medium">Visa verifikat</Link>
          </div>
        )}
      </div>

      {/* Line items */}
      {invoice.lines && invoice.lines.length > 0 && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-6">
          <div className="p-6 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">Fakturarader</h2>
          </div>
          <div className="divide-y divide-gray-100">
            {invoice.lines.map((line) => (
              <div key={line.line_id} className="p-4">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm text-gray-900">{line.description}</span>
                  <span className="text-sm font-medium text-gray-900">{formatSEK(line.line_total)} kr</span>
                </div>
                <div className="text-xs text-gray-500">
                  {parseMoney(line.quantity)} {line.unit} &times; {formatSEK(line.unit_price)} kr · {line.vat_rate}% moms
                </div>
              </div>
            ))}
          </div>

          {/* Totals */}
          <div className="p-6 border-t bg-gray-50">
            <div className="flex justify-end">
              <div className="w-72 space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Netto:</span>
                  <span className="text-gray-900">{formatSEK(invoice.subtotal)} kr</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Moms:</span>
                  <span className="text-gray-900">{formatSEK(invoice.vat_total)} kr</span>
                </div>
                {parseMoney(invoice.rounding) !== 0 && (
                  <div className="flex justify-between">
                    <span className="text-gray-600">Öresavrundning:</span>
                    <span className="text-gray-900">{formatSEK(invoice.rounding)} kr</span>
                  </div>
                )}
                <div className="flex justify-between border-t pt-2 text-base">
                  <span className="font-semibold text-gray-900">Att betala:</span>
                  <span className="font-bold text-gray-900">{formatSEK(invoice.total)} kr</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Åtgärder</h2>

        {status === "draft" && (
          <div className="flex gap-3">
            <button onClick={handleFinalize} disabled={actionLoading}
              className="px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm disabled:opacity-50">
              {actionLoading ? "Skickar..." : "Skicka faktura"}
            </button>
            <button onClick={handleDelete}
              className="px-4 py-2.5 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium text-sm">
              Radera utkast
            </button>
          </div>
        )}

        {(status === "sent" || status === "overdue") && (
          <div className="space-y-4">
            <div className="flex items-end gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Betalningsdatum</label>
                <input type="date" value={paymentDate}
                  onChange={(e) => setPaymentDate(e.target.value)}
                  className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
              </div>
              <button onClick={handlePay} disabled={actionLoading}
                className="px-4 py-2.5 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors font-medium text-sm disabled:opacity-50">
                {actionLoading ? "Registrerar..." : "Markera som betald"}
              </button>
            </div>
            <button onClick={handleCancel} disabled={actionLoading}
              className="px-4 py-2.5 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors font-medium text-sm disabled:opacity-50">
              Makulera faktura
            </button>
          </div>
        )}

        {status === "paid" && (
          <p className="text-green-600 font-medium">
            Betald {invoice.paid_at ? new Date(invoice.paid_at).toLocaleDateString("sv-SE") : ""}
          </p>
        )}

        {status === "cancelled" && (
          <p className="text-gray-500">Fakturan är makulerad</p>
        )}
      </div>
    </DashboardLayout>
  );
}
