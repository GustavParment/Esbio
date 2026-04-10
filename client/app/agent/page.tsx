"use client";

import { useState, useRef, useEffect } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";
import { agentApi } from "@/lib/api/agent";
import Link from "next/link";

interface Message {
  role: "user" | "assistant";
  content: string;
}

export default function AgentPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [conversationId, setConversationId] = useState<string>("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Handle mobile keyboard - resize container to fit visible area
  const containerRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const viewport = window.visualViewport;
    if (!viewport) return;

    const handleResize = () => {
      if (containerRef.current) {
        const offsetTop = containerRef.current.getBoundingClientRect().top;
        const availableHeight = viewport.height - offsetTop;
        containerRef.current.style.height = `${availableHeight - 16}px`;
      }
    };

    viewport.addEventListener("resize", handleResize);
    viewport.addEventListener("scroll", handleResize);
    return () => {
      viewport.removeEventListener("resize", handleResize);
      viewport.removeEventListener("scroll", handleResize);
    };
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const userMessage = input.trim();
    setInput("");
    setMessages((prev) => [...prev, { role: "user", content: userMessage }]);
    setLoading(true);

    // Add empty assistant message that we'll stream into
    setMessages((prev) => [...prev, { role: "assistant", content: "" }]);

    try {
      await agentApi.chatStream(userMessage, conversationId, {
        onText: (text) => {
          setMessages((prev) => {
            const updated = [...prev];
            const last = updated[updated.length - 1];
            if (last.role === "assistant") {
              updated[updated.length - 1] = { ...last, content: last.content + text };
            }
            return updated;
          });
        },
        onTool: (toolName) => {
          setMessages((prev) => {
            const updated = [...prev];
            const last = updated[updated.length - 1];
            if (last.role === "assistant") {
              const status = `⏳ Använder ${toolName}...\n`;
              updated[updated.length - 1] = { ...last, content: last.content + status };
            }
            return updated;
          });
        },
        onConversationId: (id) => {
          setConversationId(id);
        },
        onError: (error) => {
          setMessages((prev) => {
            const updated = [...prev];
            const last = updated[updated.length - 1];
            if (last.role === "assistant") {
              updated[updated.length - 1] = { ...last, content: "Fel: " + error };
            }
            return updated;
          });
        },
        onDone: () => {
          setLoading(false);
          // Remove tool status lines from final message
          setMessages((prev) => {
            const updated = [...prev];
            const last = updated[updated.length - 1];
            if (last.role === "assistant") {
              const cleaned = last.content.replace(/⏳ Använder .*?\.\.\.\n/g, "");
              updated[updated.length - 1] = { ...last, content: cleaned };
            }
            return updated;
          });
          inputRef.current?.focus();
        },
      });
    } catch (err) {
      setMessages((prev) => {
        const updated = [...prev];
        const last = updated[updated.length - 1];
        if (last.role === "assistant") {
          updated[updated.length - 1] = { ...last, content: "Fel: Kunde inte kommunicera med Ester AI. Försök igen senare." };
        }
        return updated;
      });
      setLoading(false);
      inputRef.current?.focus();
    }
  };

  const startNewConversation = () => {
    setMessages([]);
    setConversationId("");
    inputRef.current?.focus();
  };

  return (
    <DashboardLayout>
      <div ref={containerRef} className="flex flex-col" style={{ height: "calc(100dvh - 10rem)" }}>
        {/* Header */}
        <div className="mb-3 shrink-0">
          <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Ester AI</h1>
          <p className="text-gray-600 text-sm mt-1">Din intelligenta bokföringsassistent</p>
          <div className="grid grid-cols-2 gap-3 mt-3">
            <button
              onClick={startNewConversation}
              className="w-full py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 transition-colors font-medium text-sm"
            >
              Ny konversation
            </button>
            <Link
              href="/agent/scheduled"
              className="w-full py-3 border border-gray-300 text-gray-700 rounded-xl hover:bg-gray-50 transition-colors font-medium text-sm text-center"
            >
              Schemalagda uppgifter
            </Link>
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 min-h-0 overflow-y-auto bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-3">
          {messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-center px-2">
              <img src="/ester-banner.png" alt="Ester AI" className="w-24 h-24 rounded-full mb-4" />
              <p className="text-gray-500 text-sm max-w-md mb-6">
                Jag kan hjälpa dig att skapa verifikat, visa rapporter, kontrollera saldon och
                schemalägga återkommande bokföringar.
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2 w-full max-w-lg">
                {[
                  "Visa resultaträkning för 2025",
                  "Skapa ett konsultarvode på 50 000 kr inkl moms",
                  "Vad är saldot på konto 1930?",
                  "Schemalägga ett verifikat varje månad",
                ].map((suggestion) => (
                  <button
                    key={suggestion}
                    onClick={() => {
                      setInput(suggestion);
                      inputRef.current?.focus();
                    }}
                    className="text-left text-sm p-3 rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors text-gray-700"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map((msg, idx) => (
                <div
                  key={idx}
                  className={`flex items-end gap-2 ${msg.role === "user" ? "justify-end" : "justify-start"}`}
                >
                  {msg.role === "assistant" && (
                    <img src="/ester-banner.png" alt="Ester" className="w-7 h-7 rounded-full shrink-0" />
                  )}
                  <div
                    className={`max-w-[80%] rounded-2xl px-4 py-3 ${
                      msg.role === "user"
                        ? "bg-blue-600 text-white"
                        : "bg-gray-100 text-gray-900"
                    }`}
                  >
                    <div className="whitespace-pre-wrap text-sm">{msg.content}</div>
                  </div>
                </div>
              ))}
              {loading && (
                <div className="flex justify-start">
                  <div className="bg-gray-100 rounded-2xl px-4 py-3">
                    <div className="flex gap-1">
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: "0ms" }} />
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: "150ms" }} />
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: "300ms" }} />
                    </div>
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* Input */}
        <form onSubmit={handleSubmit} className="flex gap-2 shrink-0 pb-1">
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Skriv ett meddelande..."
            className="flex-1 min-w-0 px-3 md:px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 placeholder:text-gray-400 text-base"
            style={{ fontSize: "16px" }}
            disabled={loading}
            autoComplete="off"
          />
          <button
            type="submit"
            disabled={loading || !input.trim()}
            className="px-4 md:px-6 py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed shrink-0 text-sm md:text-base"
          >
            Skicka
          </button>
        </form>
      </div>
    </DashboardLayout>
  );
}
