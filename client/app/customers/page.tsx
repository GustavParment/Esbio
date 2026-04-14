"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { customersApi } from "@/lib/api/customers";
import { Customer } from "@/types";
import { useAuth } from "@/lib/contexts/AuthContext";
import Link from "next/link";

export default function CustomersPage() {
  const { user } = useAuth();
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");

  useEffect(() => {
    if (user) fetchCustomers();
  }, [user]);

  const fetchCustomers = async () => {
    setLoading(true);
    try {
      const data = await customersApi.getAll();
      setCustomers(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Failed to fetch customers:", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredCustomers = customers.filter((c) => {
    if (!searchTerm) return true;
    const q = searchTerm.toLowerCase();
    return (
      c.name.toLowerCase().includes(q) ||
      (c.org_number || "").toLowerCase().includes(q) ||
      (c.email || "").toLowerCase().includes(q) ||
      (c.city || "").toLowerCase().includes(q)
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
      <div className="mb-8">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Kunder</h1>
            <p className="text-gray-600 mt-2">Hantera ditt kundregister</p>
          </div>
          <Link
            href="/customers/new"
            className="px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm text-center"
          >
            + Ny kund
          </Link>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-2">
          Sök kund
        </label>
        <input
          id="search"
          type="text"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Namn, org.nr, e-post eller ort..."
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400"
        />
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {filteredCustomers.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 mb-4">
              {customers.length === 0
                ? "Inga kunder registrerade"
                : "Inga kunder matchar din sökning"}
            </p>
            {customers.length === 0 && (
              <Link
                href="/customers/new"
                className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Skapa din första kund
              </Link>
            )}
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {filteredCustomers.map((customer) => (
              <Link
                key={customer.customer_id}
                href={`/customers/${customer.customer_id}`}
                className="block p-4 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-gray-900">{customer.name}</span>
                  <span className="text-xs text-gray-500">{customer.payment_terms_days} dagar</span>
                </div>
                <div className="flex flex-wrap items-center gap-3 text-xs text-gray-500">
                  {customer.org_number && <span>{customer.org_number}</span>}
                  {customer.email && <span>{customer.email}</span>}
                  {customer.city && <span>{customer.city}</span>}
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>

      <div className="mt-6 text-sm text-gray-600">
        Visar {filteredCustomers.length} av {customers.length} kunder
      </div>
    </DashboardLayout>
  );
}
