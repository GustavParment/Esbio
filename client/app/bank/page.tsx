"use client";

import { Suspense, useEffect, useState, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { bankApi } from "@/lib/api/bank";
import { accountsApi } from "@/lib/api/accounts";
import { BankConnection, BankAccount, BankTransaction, Account, BankImportStatus } from "@/types";
import { formatSEK, parseMoney } from "@/lib/money";
import { useAuth } from "@/lib/contexts/AuthContext";

const TABS: { label: string; value: string }[] = [
  { label: "Att granska", value: "pending" },
  { label: "Alla", value: "" },
  { label: "Importerade", value: "booked" },
  { label: "Hoppade", value: "skipped" },
];

const STATUS_LABELS: Record<BankImportStatus, string> = {
  pending: "Ny",
  suggested: "Föreslagen",
  approved: "Godkänd",
  skipped: "Hoppad",
  booked: "Importerad",
};

const STATUS_COLORS: Record<BankImportStatus, string> = {
  pending: "bg-yellow-100 text-yellow-700",
  suggested: "bg-blue-100 text-blue-700",
  approved: "bg-green-100 text-green-700",
  skipped: "bg-gray-100 text-gray-500",
  booked: "bg-green-100 text-green-700",
};

function BankPageContent() {
  const { user } = useAuth();
  const searchParams = useSearchParams();
  const [connections, setConnections] = useState<BankConnection[]>([]);
  const [bankAccounts, setBankAccounts] = useState<BankAccount[]>([]);
  const [transactions, setTransactions] = useState<BankTransaction[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState<number | null>(null);
  const [activeTab, setActiveTab] = useState("pending");
  const [connectSuccess, setConnectSuccess] = useState(false);

  // Import form state
  const [importingTxnId, setImportingTxnId] = useState<number | null>(null);
  const [importAccountNo, setImportAccountNo] = useState<number>(0);
  const [importDescription, setImportDescription] = useState("");
  const [importTaxCode, setImportTaxCode] = useState<number>(0);
  const [importLoading, setImportLoading] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const [conns, accts, txns, basAccounts] = await Promise.all([
        bankApi.getConnections(),
        bankApi.getAccounts(),
        bankApi.getTransactions({ status: activeTab || undefined }),
        accountsApi.getAll(),
      ]);
      setConnections(Array.isArray(conns) ? conns : []);
      setBankAccounts(Array.isArray(accts) ? accts : []);
      setTransactions(Array.isArray(txns) ? txns : []);
      setAccounts(Array.isArray(basAccounts) ? basAccounts : []);
    } catch (error) {
      console.error("Failed to fetch bank data:", error);
    } finally {
      setLoading(false);
    }
  }, [activeTab]);

  useEffect(() => {
    if (user) fetchData();
  }, [user, fetchData]);

  useEffect(() => {
    if (searchParams.get("connected") === "true") {
      setConnectSuccess(true);
      setTimeout(() => setConnectSuccess(false), 5000);
    }
  }, [searchParams]);

  const handleConnect = async () => {
    try {
      const { url } = await bankApi.connect();
      window.location.href = url;
    } catch (error) {
      console.error("Failed to initiate bank connection:", error);
    }
  };

  const handleSync = async (connectionId: number) => {
    setSyncing(connectionId);
    try {
      const result = await bankApi.syncTransactions(connectionId);
      alert(`Synkade ${result.new_count} nya transaktioner`);
      fetchData();
    } catch (error) {
      console.error("Failed to sync:", error);
      alert("Kunde inte synka transaktioner. Kontrollera bankkopplingen.");
    } finally {
      setSyncing(null);
    }
  };

  const handleDisconnect = async (connectionId: number) => {
    if (!confirm("Vill du koppla bort denna bank? Alla importerade transaktioner behålls.")) return;
    try {
      await bankApi.disconnect(connectionId);
      fetchData();
    } catch (error) {
      console.error("Failed to disconnect:", error);
    }
  };

  const handleMapAccount = async (bankAccountId: number, accountNo: number) => {
    try {
      await bankApi.mapAccount(bankAccountId, accountNo);
      fetchData();
    } catch (error) {
      console.error("Failed to map account:", error);
    }
  };

  const startImport = (txn: BankTransaction) => {
    setImportingTxnId(txn.bank_transaction_id);
    setImportAccountNo(txn.suggested_account_no || 0);
    setImportDescription(txn.suggested_description || txn.description_display || txn.description_original);
    setImportTaxCode(0);
  };

  const handleImport = async () => {
    if (!importingTxnId || !importAccountNo || !importDescription) return;
    setImportLoading(true);
    try {
      await bankApi.importTransaction(importingTxnId, {
        account_no: importAccountNo,
        description: importDescription,
        tax_code: importTaxCode,
      });
      setImportingTxnId(null);
      fetchData();
    } catch (error) {
      console.error("Failed to import transaction:", error);
      alert("Kunde inte importera transaktionen.");
    } finally {
      setImportLoading(false);
    }
  };

  const handleSkip = async (txnId: number) => {
    try {
      await bankApi.skipTransaction(txnId);
      fetchData();
    } catch (error) {
      console.error("Failed to skip transaction:", error);
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500" />
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="mb-8">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Bank</h1>
            <p className="text-gray-600 mt-2">Koppla din bank och importera transaktioner</p>
          </div>
          <button
            onClick={handleConnect}
            className="px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm"
          >
            + Koppla bank
          </button>
        </div>
      </div>

      {connectSuccess && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
          <p className="text-green-700">Banken kopplades framgangsrikt! Klicka Synka for att hamta transaktioner.</p>
        </div>
      )}

      {/* Connected banks */}
      {connections.length > 0 && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Kopplade banker</h2>
          <div className="space-y-4">
            {connections.map((conn) => {
              const connAccounts = bankAccounts.filter((a) => a.connection_id === conn.connection_id);
              return (
                <div key={conn.connection_id} className="border border-gray-100 rounded-lg p-4">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
                    <div>
                      <h3 className="font-medium text-gray-900">{conn.bank_name}</h3>
                      <p className="text-sm text-gray-500">
                        {conn.last_synced_at
                          ? `Senast synkad: ${new Date(conn.last_synced_at).toLocaleString("sv-SE")}`
                          : "Aldrig synkad"}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleSync(conn.connection_id)}
                        disabled={syncing === conn.connection_id}
                        className="px-3 py-1.5 bg-blue-50 text-blue-700 rounded-lg hover:bg-blue-100 transition-colors text-sm font-medium disabled:opacity-50"
                      >
                        {syncing === conn.connection_id ? "Synkar..." : "Synka"}
                      </button>
                      <button
                        onClick={() => handleDisconnect(conn.connection_id)}
                        className="px-3 py-1.5 bg-red-50 text-red-700 rounded-lg hover:bg-red-100 transition-colors text-sm font-medium"
                      >
                        Koppla bort
                      </button>
                    </div>
                  </div>

                  {/* Bank accounts with mapping */}
                  {connAccounts.length > 0 && (
                    <div className="mt-3 space-y-2">
                      {connAccounts.map((ba) => (
                        <div key={ba.bank_account_id} className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4 bg-gray-50 rounded-lg p-3">
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 flex-wrap">
                              <span className="font-medium text-gray-900">{ba.account_name}</span>
                              {ba.account_type && (
                                <span className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full">{ba.account_type}</span>
                              )}
                              {ba.balance_amount && (
                                <span className="font-medium text-gray-700">{formatSEK(ba.balance_amount)}</span>
                              )}
                            </div>
                            <div className="flex items-center gap-2 text-xs text-gray-400 mt-0.5">
                              {ba.holder_name && <span>{ba.holder_name}</span>}
                              {ba.holder_name && ba.iban && <span>·</span>}
                              {ba.iban && <span className="truncate">{ba.iban}</span>}
                            </div>
                          </div>
                          <div className="flex items-center gap-2 shrink-0">
                            <label className="text-sm text-gray-500 whitespace-nowrap">BAS-konto:</label>
                            <select
                              value={ba.mapped_account_no || ""}
                              onChange={(e) =>
                                e.target.value
                                  ? handleMapAccount(ba.bank_account_id, parseInt(e.target.value))
                                  : null
                              }
                              className="border border-gray-300 rounded-lg px-2 py-1 text-sm bg-white w-full sm:w-auto"
                            >
                              <option value="">Välj konto...</option>
                              {accounts
                                .filter((a) => a.account_group === 1 || a.account_group === 2)
                                .map((a) => (
                                  <option key={a.account_no} value={a.account_no}>
                                    {a.account_no} {a.account_name}
                                  </option>
                                ))}
                            </select>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* No connections state */}
      {connections.length === 0 && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center mb-6">
          <svg className="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
          </svg>
          <h3 className="text-lg font-medium text-gray-900 mb-2">Ingen bank kopplad</h3>
          <p className="text-gray-500 mb-4">Koppla din bank for att automatiskt importera transaktioner.</p>
          <button
            onClick={handleConnect}
            className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Koppla bank
          </button>
        </div>
      )}

      {/* Transactions */}
      {connections.length > 0 && (
        <>
          <div className="flex gap-1 mb-6 bg-gray-100 rounded-lg p-1 overflow-x-auto">
            {TABS.map((tab) => (
              <button
                key={tab.value}
                onClick={() => setActiveTab(tab.value)}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors whitespace-nowrap ${
                  activeTab === tab.value
                    ? "bg-white text-gray-900 shadow-sm"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
            {transactions.length === 0 ? (
              <div className="text-center py-12">
                <p className="text-gray-500">
                  {activeTab === "pending"
                    ? "Inga transaktioner att granska. Klicka Synka for att hamta nya."
                    : "Inga transaktioner med denna status."}
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Datum</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Beskrivning</th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Belopp</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Konto</th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Atgarder</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {transactions.map((txn) => {
                      const amount = parseMoney(txn.amount);
                      const isPositive = amount >= 0;
                      const isImporting = importingTxnId === txn.bank_transaction_id;

                      return (
                        <tr key={txn.bank_transaction_id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 text-sm text-gray-900 whitespace-nowrap">{txn.booked_date}</td>
                          <td className="px-4 py-3 text-sm text-gray-900">
                            <div>{txn.description_display || txn.description_original}</div>
                            {txn.merchant_name && (
                              <div className="text-xs text-gray-400">{txn.merchant_name}</div>
                            )}
                          </td>
                          <td className={`px-4 py-3 text-sm text-right font-medium whitespace-nowrap ${isPositive ? "text-green-600" : "text-red-600"}`}>
                            {formatSEK(txn.amount)}
                          </td>
                          <td className="px-4 py-3">
                            <span className={`inline-flex px-2 py-1 text-xs rounded-full font-medium ${STATUS_COLORS[txn.import_status]}`}>
                              {STATUS_LABELS[txn.import_status]}
                            </span>
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">
                            {txn.suggested_account_no && (
                              <span className="text-blue-600">
                                {txn.suggested_account_no}{" "}
                                {accounts.find((a) => a.account_no === txn.suggested_account_no)?.account_name || ""}
                              </span>
                            )}
                            {txn.voucher_id && (
                              <a href={`/vouchers/${txn.voucher_id}`} className="text-blue-600 hover:underline">
                                Verifikat #{txn.voucher_id}
                              </a>
                            )}
                          </td>
                          <td className="px-4 py-3 text-right">
                            {(txn.import_status === "pending" || txn.import_status === "suggested") && !isImporting && (
                              <div className="flex gap-2 justify-end">
                                <button
                                  onClick={() => startImport(txn)}
                                  className="px-3 py-1 bg-green-50 text-green-700 rounded-lg hover:bg-green-100 transition-colors text-sm font-medium"
                                >
                                  Importera
                                </button>
                                <button
                                  onClick={() => handleSkip(txn.bank_transaction_id)}
                                  className="px-3 py-1 bg-gray-50 text-gray-600 rounded-lg hover:bg-gray-100 transition-colors text-sm font-medium"
                                >
                                  Hoppa
                                </button>
                              </div>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* Import form modal */}
          {importingTxnId && (
            <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setImportingTxnId(null)}>
              <div className="bg-white rounded-xl shadow-xl p-6 w-full max-w-md mx-4" onClick={(e) => e.stopPropagation()}>
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Importera transaktion</h3>

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Motkonto</label>
                    <select
                      value={importAccountNo}
                      onChange={(e) => setImportAccountNo(parseInt(e.target.value))}
                      className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                    >
                      <option value={0}>Valj konto...</option>
                      {accounts.map((a) => (
                        <option key={a.account_no} value={a.account_no}>
                          {a.account_no} {a.account_name}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Beskrivning</label>
                    <input
                      type="text"
                      value={importDescription}
                      onChange={(e) => setImportDescription(e.target.value)}
                      className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Momskod</label>
                    <select
                      value={importTaxCode}
                      onChange={(e) => setImportTaxCode(parseInt(e.target.value))}
                      className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                    >
                      <option value={0}>0% (Momsfri)</option>
                      <option value={6}>6%</option>
                      <option value={12}>12%</option>
                      <option value={25}>25%</option>
                    </select>
                  </div>

                  <div className="flex gap-3 pt-2">
                    <button
                      onClick={handleImport}
                      disabled={!importAccountNo || !importDescription || importLoading}
                      className="flex-1 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors font-medium text-sm disabled:opacity-50"
                    >
                      {importLoading ? "Importerar..." : "Skapa verifikat"}
                    </button>
                    <button
                      onClick={() => setImportingTxnId(null)}
                      className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors font-medium text-sm"
                    >
                      Avbryt
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </DashboardLayout>
  );
}

export default function BankPage() {
  return (
    <Suspense>
      <BankPageContent />
    </Suspense>
  );
}
