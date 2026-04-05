"use client";

import DashboardLayout from "@/components/layout/DashboardLayout";
import Link from "next/link";

export default function ReportsPage() {
  const reportTypes = [
    {
      name: "Resultaträkning",
      description: "Visa företagets intäkter och kostnader",
      icon: "📊",
      color: "bg-blue-500",
      href: "/reports/income-statement",
    },
    {
      name: "Balansräkning",
      description: "Visa företagets tillgångar och skulder",
      icon: "⚖️",
      color: "bg-green-500",
      href: "/reports/balance-sheet",
    },
    {
      name: "Kontoutdrag",
      description: "Visa transaktioner för specifika konton",
      icon: "📋",
      color: "bg-purple-500",
      href: "/accounts",
    },
    {
      name: "Huvudbok",
      description: "Visa alla bokförda verifikat",
      icon: "📚",
      color: "bg-orange-500",
      href: "/vouchers",
    },
    {
      name: "Momsrapport",
      description: "Visa momsunderlag och beräkningar",
      icon: "💶",
      color: "bg-red-500",
      href: "/reports/vat",
    },
    {
      name: "SIE4-export",
      description: "Exportera bokföring för Skatteverket",
      icon: "📁",
      color: "bg-gray-600",
      href: "/reports/sie",
    },
  ];

  return (
    <DashboardLayout>
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Rapporter</h1>
        <p className="text-gray-600 mt-2">Visa och exportera finansiella rapporter</p>
      </div>

      {/* Reports grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {reportTypes.map((report) => {
          const content = (
            <>
              <div className={`${report.color} w-12 h-12 rounded-lg flex items-center justify-center text-2xl mb-4`}>
                {report.icon}
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">{report.name}</h3>
              <p className="text-sm text-gray-600">{report.description}</p>
              <div className={`mt-4 text-sm font-medium ${report.href ? 'text-blue-600' : 'text-gray-400'}`}>
                Visa rapport →
              </div>
            </>
          );

          if (report.href) {
            return (
              <Link
                key={report.name}
                href={report.href}
                className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow block"
              >
                {content}
              </Link>
            );
          }

          return (
            <div
              key={report.name}
              className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 opacity-60 cursor-not-allowed"
            >
              {content}
            </div>
          );
        })}
      </div>

    </DashboardLayout>
  );
}
