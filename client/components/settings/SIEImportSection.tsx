"use client";

import { useState } from "react";
import { reportsApi, SIEImportSummary } from "@/lib/api/reports";

export default function SIEImportSection() {
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<SIEImportSummary | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [doneMessage, setDoneMessage] = useState("");

  const reset = () => {
    setFile(null);
    setPreview(null);
    setError("");
    setDoneMessage("");
  };

  const handlePreview = async () => {
    if (!file) return;
    setBusy(true);
    setError("");
    setDoneMessage("");
    try {
      const summary = await reportsApi.uploadSIE(file, true);
      setPreview(summary);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kunde inte läsa filen");
    } finally {
      setBusy(false);
    }
  };

  const handleImport = async () => {
    if (!file) return;
    if (!confirm("Importera verifikat och konton från filen? Detta går inte att ångra automatiskt.")) return;
    setBusy(true);
    setError("");
    try {
      const result = await reportsApi.uploadSIE(file, false);
      setPreview(result);
      const r = result.imported;
      setDoneMessage(
        r
          ? `Klart! ${r.vouchers_created} verifikat och ${r.accounts_created} konton importerade (${r.vouchers_skipped} verifikat hoppades över).`
          : "Klart!"
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Importen misslyckades");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-2">Importera bokföring (SIE)</h2>
      <p className="text-sm text-gray-500 mb-4">
        Importera verifikat och kontoplan från Fortnox, Visma eller annat bokföringssystem
        via SIE 4-fil (.se). Förhandsgranska först — inget skrivs förrän du bekräftar.
      </p>

      <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center mb-4">
        <input
          type="file"
          accept=".se,.si,.sie"
          onChange={(e) => {
            setFile(e.target.files?.[0] ?? null);
            setPreview(null);
            setDoneMessage("");
            setError("");
          }}
          className="block text-sm text-gray-700 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
        />
        <button
          onClick={handlePreview}
          disabled={!file || busy}
          className="px-4 py-2 bg-gray-100 text-gray-800 rounded-lg hover:bg-gray-200 transition-colors text-sm font-medium disabled:opacity-50"
        >
          {busy && !preview ? "Läser..." : "Förhandsgranska"}
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-700 text-sm">{error}</p>
        </div>
      )}

      {doneMessage && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg">
          <p className="text-green-700 text-sm">{doneMessage}</p>
        </div>
      )}

      {preview && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
            <KV label="Företag" value={preview.company_name || "—"} />
            <KV label="Org.nummer" value={preview.org_number || "—"} />
            <KV label="Räkenskapsår" value={preview.fiscal_from && preview.fiscal_to ? `${preview.fiscal_from} — ${preview.fiscal_to}` : "—"} />
            <KV label="Kontoplan" value={preview.chart_type || "—"} />
            <KV label="Konton i filen" value={`${preview.accounts_in_file} (${preview.accounts_to_add} nya)`} />
            <KV label="Verifikat i filen" value={String(preview.vouchers_in_file)} />
            <KV label="Obalanserade verifikat" value={String(preview.unbalanced_count)} highlight={preview.unbalanced_count > 0} />
            <KV label="Teckenkodning" value={preview.encoding} />
          </div>

          {preview.warnings && preview.warnings.length > 0 && (
            <details className="mt-4">
              <summary className="text-sm font-medium text-orange-700 cursor-pointer">
                {preview.warnings.length} varning(ar) — visa
              </summary>
              <ul className="mt-2 text-xs text-orange-700 list-disc pl-5 space-y-1 max-h-40 overflow-auto">
                {preview.warnings.map((w, i) => (
                  <li key={i}>{w}</li>
                ))}
              </ul>
            </details>
          )}

          {!preview.imported && (
            <div className="mt-4 flex flex-wrap gap-2">
              <button
                onClick={handleImport}
                disabled={busy}
                className="px-5 py-2.5 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors text-sm font-medium disabled:opacity-50"
              >
                {busy ? "Importerar..." : `Importera ${preview.vouchers_in_file - preview.unbalanced_count} verifikat`}
              </button>
              <button
                onClick={reset}
                disabled={busy}
                className="px-5 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors text-sm font-medium"
              >
                Avbryt
              </button>
            </div>
          )}

          {preview.imported && (
            <button
              onClick={reset}
              className="mt-4 px-5 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors text-sm font-medium"
            >
              Importera en till fil
            </button>
          )}
        </div>
      )}
    </div>
  );
}

function KV({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div>
      <p className="text-xs text-gray-500">{label}</p>
      <p className={`font-medium ${highlight ? "text-red-700" : "text-gray-900"}`}>{value}</p>
    </div>
  );
}
