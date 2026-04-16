"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { bankApi } from "@/lib/api/bank";

function BankCallbackContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const [error, setError] = useState("");
  const [processing, setProcessing] = useState(true);

  useEffect(() => {
    const state = searchParams.get("state");
    const code = searchParams.get("code");
    const errorParam = searchParams.get("error");

    if (errorParam) {
      setError("Bankkopplingen avbröts eller misslyckades.");
      setProcessing(false);
      return;
    }

    if (!state || !code) {
      setError("Saknar state eller code från Tink.");
      setProcessing(false);
      return;
    }

    bankApi
      .handleCallback(state, code)
      .then(() => {
        window.location.href = "/bank?connected=true";
      })
      .catch((err) => {
        console.error("Bank callback error:", err);
        setError("Kunde inte koppla banken. Försök igen.");
        setProcessing(false);
      });
  }, [searchParams, router]);

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="max-w-lg mx-auto text-center">
        {processing && !error && (
          <div>
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Kopplar din bank...</h2>
            <p className="text-gray-500">Vänta medan vi hämtar dina konton.</p>
          </div>
        )}
        {error && (
          <div>
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
              <p className="text-red-700">{error}</p>
            </div>
            <button
              onClick={() => router.push("/bank")}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors"
            >
              Tillbaka till Bank
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default function BankCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-gray-50 flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto" />
        </div>
      }
    >
      <BankCallbackContent />
    </Suspense>
  );
}
