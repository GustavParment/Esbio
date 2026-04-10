import { apiClient } from "./client";

export interface ChatResponse {
  response: string;
  conversation_id: string;
}

export interface AgentMessage {
  message_id: number;
  user_id: number;
  conversation_id: string;
  role: "user" | "assistant";
  content: string;
  created_at: string;
}

export interface ScheduledTask {
  task_id: number;
  user_id: number;
  description: string;
  prompt: string;
  template_voucher_id: number | null;
  day_of_month: number;
  active: boolean;
  last_run_at: string | null;
  next_run_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface AgentUsageSummary {
  total_prompt_tokens: number;
  total_completion_tokens: number;
  total_tokens: number;
  request_count: number;
  estimated_cost_sek: number;
  month_prompt_tokens: number;
  month_completion_tokens: number;
  month_total_tokens: number;
  month_request_count: number;
  month_estimated_cost_sek: number;
}

export interface StreamCallbacks {
  onText: (text: string) => void;
  onTool: (toolName: string) => void;
  onConversationId: (id: string) => void;
  onError: (error: string) => void;
  onDone: () => void;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "https://api.esbio.se/api/v1";

export const agentApi = {
  chat: async (message: string, conversationId?: string): Promise<ChatResponse> => {
    return apiClient.post<ChatResponse>("/agent/chat", {
      message,
      conversation_id: conversationId || "",
    });
  },

  chatStream: async (message: string, conversationId: string, callbacks: StreamCallbacks): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/agent/stream`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message, conversation_id: conversationId || "" }),
    });

    if (!response.ok || !response.body) {
      callbacks.onError("Kunde inte kommunicera med Ester AI.");
      return;
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";

      let currentEvent = "";
      for (const line of lines) {
        if (line.startsWith("event: ")) {
          currentEvent = line.slice(7);
        } else if (line.startsWith("data: ")) {
          const data = line.slice(6);
          switch (currentEvent) {
            case "text":
              callbacks.onText(data);
              break;
            case "tool":
              callbacks.onTool(data);
              break;
            case "conversation_id":
              callbacks.onConversationId(data);
              break;
            case "error":
              callbacks.onError(data);
              break;
            case "done":
              callbacks.onDone();
              break;
          }
        }
      }
    }
    callbacks.onDone();
  },

  getMessages: async (conversationId: string): Promise<AgentMessage[]> => {
    return apiClient.get<AgentMessage[]>(`/agent/messages/${conversationId}`);
  },

  getScheduledTasks: async (): Promise<ScheduledTask[]> => {
    return apiClient.get<ScheduledTask[]>("/agent/tasks");
  },

  toggleTask: async (taskId: number): Promise<ScheduledTask> => {
    return apiClient.put<ScheduledTask>(`/agent/tasks/${taskId}/toggle`, {});
  },

  deleteTask: async (taskId: number): Promise<void> => {
    return apiClient.delete<void>(`/agent/tasks/${taskId}`);
  },

  getUsage: async (): Promise<AgentUsageSummary> => {
    return apiClient.get<AgentUsageSummary>("/agent/usage");
  },
};
