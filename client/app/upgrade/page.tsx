"use client";

import { useAuth } from "@/lib/contexts/AuthContext";
import { useCompany } from "@/lib/contexts/CompanyContext";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { stripeApi } from "@/lib/api/stripe";

const plans = [
  {
    id: "starter",
    name: "Starter",
    price: 199,
    description: "Perfekt för enskild firma och frilansare.",
    features: [
      "Upp till 100 transaktioner/mån",
      "AI-kontering",
      "Kvittoskanning",
      "Momsrapporter",
      "5 fakturor/mån",
    ],
    featured: false,
    priceId: "price_1TKFEmF27XxV0OojLLuaOAfg",
  },
  {
    id: "growth",
    name: "Tillväxt",
    price: 399,
    description: "För aktiebolag och växande verksamheter.",
    features: [
      "Obegränsade transaktioner",
      "AI-kontering med inlärning",
      "Kvittoskanning",
      "Momsrapporter",
      "Obegränsade fakturor",
      "Årsredovisning ingår",
      "Prioriterad support",
    ],
    featured: true,
    priceId: "price_1TKFFGF27XxV0Ooj3uaT41ho",
  },
];

export default function UpgradePage() {
  const { user, loading: authLoading } = useAuth();
  const { selectedCompany } = useCompany();
  const router = useRouter();
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!authLoading && !user) {
      router.push("/auth/login");
    }
  }, [user, authLoading, router]);

  const handleSelectPlan = async (planId: string, priceId: string) => {
    setSelectedPlan(planId);
    setProcessing(true);
    setError("");

    try {
      const { url } = await stripeApi.createCheckoutSession(priceId);
      window.location.href = url;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte starta betalning");
      setProcessing(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="inline-block px-4 py-1.5 bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400 rounded-full text-sm font-medium mb-4">
            Din provperiod har gått ut
          </div>
          <h1 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-3">
            Uppgradera för att fortsätta
          </h1>
          <p className="text-gray-600 dark:text-gray-400 text-lg max-w-xl mx-auto">
            Din 30-dagars provperiod{selectedCompany ? ` för ${selectedCompany.company_name}` : ""} har löpt ut.
            Välj en plan för att fortsätta använda Esbio.
          </p>
        </div>

        {/* Error */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-center">
            <p className="text-red-700 dark:text-red-400">{error}</p>
          </div>
        )}

        {/* Plans */}
        <div className="grid md:grid-cols-2 gap-6 mb-12">
          {plans.map((plan) => (
            <div
              key={plan.id}
              className={`relative bg-white dark:bg-gray-800 rounded-2xl shadow-sm border-2 p-8 transition-all ${
                plan.featured
                  ? "border-blue-500 dark:border-blue-400"
                  : "border-gray-200 dark:border-gray-700"
              }`}
            >
              {plan.featured && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-4 py-1 bg-blue-600 text-white text-xs font-semibold rounded-full">
                  Mest populär
                </div>
              )}

              <div className="mb-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{plan.name}</h3>
                <div className="mt-2 flex items-baseline gap-1">
                  <span className="text-4xl font-bold text-gray-900 dark:text-white">{plan.price}</span>
                  <span className="text-gray-500 dark:text-gray-400">kr/mån</span>
                </div>
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">{plan.description}</p>
              </div>

              <ul className="space-y-3 mb-8">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <svg className="w-5 h-5 text-green-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                    {feature}
                  </li>
                ))}
              </ul>

              <button
                onClick={() => handleSelectPlan(plan.id, plan.priceId)}
                disabled={processing}
                className={`w-full py-3 px-6 rounded-xl font-medium transition-colors ${
                  plan.featured
                    ? "bg-blue-600 text-white hover:bg-blue-700"
                    : "bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-200"
                } disabled:opacity-50`}
              >
                {processing && selectedPlan === plan.id ? "Bearbetar..." : `Välj ${plan.name}`}
              </button>
            </div>
          ))}
        </div>

        {/* Contact */}
        <div className="text-center bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-8">
          <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Behöver du Enterprise?</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
            För redovisningsbyråer och bolag med komplexa behov — kontakta oss för en skräddarsydd offert.
          </p>
          <a
            href="mailto:info@esbio.se"
            className="inline-block px-6 py-2.5 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-sm font-medium"
          >
            Kontakta oss
          </a>
        </div>

        {/* Footer note */}
        <p className="text-center text-xs text-gray-400 dark:text-gray-500 mt-8">
          Alla priser exklusive moms. Du kan avbryta när som helst.
        </p>
      </div>
    </div>
  );
}
