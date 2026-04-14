"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { customersApi } from "@/lib/api/customers";
import { Customer } from "@/types";
import Link from "next/link";

export default function CustomerDetailPage() {
  const params = useParams();
  const router = useRouter();
  const customerId = Number(params.id);

  const [customer, setCustomer] = useState<Customer | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const [form, setForm] = useState({
    name: "", org_number: "", vat_number: "", address_line1: "", address_line2: "",
    postal_code: "", city: "", country: "", email: "", phone: "",
    payment_terms_days: 30, notes: "",
  });

  useEffect(() => {
    const fetchCustomer = async () => {
      try {
        const data = await customersApi.getById(customerId);
        setCustomer(data);
        setForm({
          name: data.name || "", org_number: data.org_number || "",
          vat_number: data.vat_number || "", address_line1: data.address_line1 || "",
          address_line2: data.address_line2 || "", postal_code: data.postal_code || "",
          city: data.city || "", country: data.country || "Sverige",
          email: data.email || "", phone: data.phone || "",
          payment_terms_days: data.payment_terms_days, notes: data.notes || "",
        });
      } catch (err) {
        setError("Kunde inte hitta kunden");
      } finally {
        setLoading(false);
      }
    };
    if (!isNaN(customerId)) fetchCustomer();
  }, [customerId]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSave = async () => {
    setSaving(true);
    setError("");
    try {
      const updated = await customersApi.update(customerId, {
        ...form, payment_terms_days: Number(form.payment_terms_days) || 30,
      });
      setCustomer(updated);
      setEditing(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte uppdatera");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Radera kunden? Detta kan inte ångras.")) return;
    try {
      await customersApi.delete(customerId);
      router.push("/customers");
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

  if (!customer) {
    return (
      <DashboardLayout>
        <div className="text-center py-12">
          <p className="text-gray-500 mb-4">{error || "Kund hittades inte"}</p>
          <Link href="/customers" className="text-blue-600 hover:text-blue-700 font-medium">
            &larr; Tillbaka till kunder
          </Link>
        </div>
      </DashboardLayout>
    );
  }

  const Field = ({ label, value }: { label: string; value: string }) => (
    <div>
      <p className="text-sm text-gray-500 mb-1">{label}</p>
      <p className="font-medium text-gray-900">{value || "-"}</p>
    </div>
  );

  return (
    <DashboardLayout>
      <div className="mb-8">
        <Link href="/customers" className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block">
          &larr; Tillbaka till kunder
        </Link>
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-gray-900">{customer.name}</h1>
          <div className="flex gap-3">
            {!editing ? (
              <>
                <button onClick={() => setEditing(true)}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
                  Redigera
                </button>
                <button onClick={handleDelete}
                  className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors text-sm font-medium">
                  Radera
                </button>
              </>
            ) : (
              <>
                <button onClick={() => setEditing(false)}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors text-sm font-medium">
                  Avbryt
                </button>
                <button onClick={handleSave} disabled={saving}
                  className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm font-medium disabled:opacity-50">
                  {saving ? "Sparar..." : "Spara"}
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">{error}</div>
      )}

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Kunduppgifter</h2>
        {editing ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[
              { name: "name", label: "Namn" }, { name: "org_number", label: "Org.nr" },
              { name: "vat_number", label: "Momsreg.nr" }, { name: "email", label: "E-post" },
              { name: "phone", label: "Telefon" },
              { name: "address_line1", label: "Adress 1" }, { name: "address_line2", label: "Adress 2" },
              { name: "postal_code", label: "Postnummer" }, { name: "city", label: "Ort" },
              { name: "country", label: "Land" },
            ].map((f) => (
              <div key={f.name}>
                <label className="block text-sm font-medium text-gray-700 mb-1">{f.label}</label>
                <input name={f.name} value={(form as Record<string, string | number>)[f.name] || ""}
                  onChange={handleChange}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
              </div>
            ))}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Betalningsvillkor (dagar)</label>
              <input name="payment_terms_days" type="number" value={form.payment_terms_days}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Field label="Org.nummer" value={customer.org_number || ""} />
            <Field label="Momsreg.nr" value={customer.vat_number || ""} />
            <Field label="E-post" value={customer.email || ""} />
            <Field label="Telefon" value={customer.phone || ""} />
            <Field label="Betalningsvillkor" value={`${customer.payment_terms_days} dagar`} />
            <Field label="Ort" value={customer.city || ""} />
            <Field label="Adress" value={[customer.address_line1, customer.address_line2].filter(Boolean).join(", ") || ""} />
            <Field label="Postnummer" value={customer.postal_code || ""} />
            <Field label="Land" value={customer.country || ""} />
          </div>
        )}
      </div>

      {!editing && customer.notes && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-2">Anteckningar</h2>
          <p className="text-gray-700 whitespace-pre-wrap">{customer.notes}</p>
        </div>
      )}

      {editing && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-2">Anteckningar</h2>
          <textarea name="notes" value={form.notes} onChange={handleChange} rows={3}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900" />
        </div>
      )}
    </DashboardLayout>
  );
}
