"use client";

import { useAuth } from "@/lib/contexts/AuthContext";
import { useCompany } from "@/lib/contexts/CompanyContext";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading: authLoading } = useAuth();
  const { selectedCompany, loading: compLoading } = useCompany();
  const router = useRouter();

  useEffect(() => {
    if (!authLoading && !user) {
      router.push("/auth/login");
    }
  }, [user, authLoading, router]);

  useEffect(() => {
    if (!authLoading && user && !compLoading && !selectedCompany) {
      router.push("/companies");
    }
  }, [user, authLoading, selectedCompany, compLoading, router]);

  if (authLoading || compLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!user || !selectedCompany) {
    return null;
  }

  return <>{children}</>;
}
