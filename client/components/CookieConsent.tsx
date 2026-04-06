"use client";

import { useState, useEffect } from "react";

export default function CookieConsent() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const consent = localStorage.getItem("cookie-consent");
    if (!consent) {
      setVisible(true);
    }
  }, []);

  const accept = () => {
    localStorage.setItem("cookie-consent", "accepted");
    setVisible(false);
  };

  if (!visible) return null;

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 p-4 bg-gray-900 border-t border-gray-700 shadow-lg">
      <div className="max-w-4xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-4">
        <p className="text-sm text-gray-300 text-center sm:text-left">
          Vi använder cookies för att hålla dig inloggad och förbättra din upplevelse.
          Genom att fortsätta använda Esbio godkänner du vår användning av cookies.
        </p>
        <button
          onClick={accept}
          className="shrink-0 px-6 py-2 bg-[var(--brand)] hover:bg-[var(--brand-hover)] text-white font-medium rounded-lg transition-colors text-sm"
        >
          Jag godkänner
        </button>
      </div>
    </div>
  );
}
