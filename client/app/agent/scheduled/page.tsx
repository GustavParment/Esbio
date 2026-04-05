"use client";

import { useEffect, useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { agentApi, ScheduledTask } from "@/lib/api/agent";
import Link from "next/link";

export default function ScheduledTasksPage() {
  const [tasks, setTasks] = useState<ScheduledTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchTasks = async () => {
    try {
      const data = await agentApi.getScheduledTasks();
      setTasks(Array.isArray(data) ? data : []);
    } catch (err) {
      setError("Kunde inte hämta schemalagda uppgifter");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  const handleToggle = async (taskId: number) => {
    try {
      const updated = await agentApi.toggleTask(taskId);
      setTasks((prev) =>
        prev.map((t) => (t.task_id === taskId ? updated : t))
      );
    } catch {
      setError("Kunde inte uppdatera uppgift");
    }
  };

  const handleDelete = async (taskId: number) => {
    if (!confirm("Vill du ta bort denna schemalagda uppgift?")) return;
    try {
      await agentApi.deleteTask(taskId);
      setTasks((prev) => prev.filter((t) => t.task_id !== taskId));
    } catch {
      setError("Kunde inte ta bort uppgift");
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
      {/* Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <Link
            href="/agent"
            className="text-blue-600 hover:text-blue-700 font-medium mb-4 inline-block"
          >
            &larr; Tillbaka till Ester AI
          </Link>
          <h1 className="text-3xl font-bold text-gray-900">Schemalagda uppgifter</h1>
          <p className="text-gray-600 mt-2">
            Verifikat och uppgifter som skapas automatiskt varje månad
          </p>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-700">{error}</p>
        </div>
      )}

      {tasks.length === 0 ? (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
          <div className="text-5xl mb-4">📅</div>
          <p className="text-gray-500 mb-4">Inga schemalagda uppgifter</p>
          <p className="text-gray-400 text-sm max-w-md mx-auto">
            Be Ester att schemalägga ett verifikat, t.ex. &quot;Skapa konsultarvode varje
            månad den 30:e&quot;
          </p>
          <Link
            href="/agent"
            className="inline-block mt-4 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
          >
            Gå till assistent
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {tasks.map((task) => (
            <div
              key={task.task_id}
              className={`bg-white rounded-xl shadow-sm border border-gray-200 p-6 ${
                !task.active ? "opacity-60" : ""
              }`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="text-lg font-semibold text-gray-900">
                      {task.description}
                    </h3>
                    <span
                      className={`px-2 py-1 rounded-full text-xs font-medium ${
                        task.active
                          ? "bg-green-100 text-green-700"
                          : "bg-gray-100 text-gray-500"
                      }`}
                    >
                      {task.active ? "Aktiv" : "Pausad"}
                    </span>
                  </div>

                  <p className="text-sm text-gray-600 mb-3">{task.prompt}</p>

                  <div className="flex gap-6 text-sm text-gray-500">
                    <div>
                      Körs den <span className="font-medium text-gray-700">{task.day_of_month}:e</span> varje månad
                    </div>
                    {task.last_run_at && (
                      <div>
                        Senast körd:{" "}
                        <span className="font-medium text-gray-700">
                          {new Date(task.last_run_at).toLocaleDateString("sv-SE")}
                        </span>
                      </div>
                    )}
                    {task.next_run_at && (
                      <div>
                        Nästa körning:{" "}
                        <span className="font-medium text-gray-700">
                          {new Date(task.next_run_at).toLocaleDateString("sv-SE")}
                        </span>
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex gap-2 ml-4">
                  <button
                    onClick={() => handleToggle(task.task_id)}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                      task.active
                        ? "border border-yellow-400 text-yellow-700 hover:bg-yellow-50"
                        : "border border-green-400 text-green-700 hover:bg-green-50"
                    }`}
                  >
                    {task.active ? "Pausa" : "Aktivera"}
                  </button>
                  <button
                    onClick={() => handleDelete(task.task_id)}
                    className="px-3 py-1.5 rounded-lg text-sm font-medium border border-red-300 text-red-600 hover:bg-red-50 transition-colors"
                  >
                    Ta bort
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </DashboardLayout>
  );
}
