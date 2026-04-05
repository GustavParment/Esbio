"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { useAuth } from "@/lib/contexts/AuthContext";
import { useTheme } from "@/lib/contexts/ThemeContext";
import { useState } from "react";

const navigation = [
  { name: "Dashboard", href: "/dashboard", icon: "📊" },
  { name: "Verifikat", href: "/vouchers", icon: "📝" },
  { name: "Konton", href: "/accounts", icon: "💰" },
  { name: "Rapporter", href: "/reports", icon: "📈" },
  { name: "Ester AI", href: "/agent", icon: "ester" },
];

const adminNavigation = [
  { name: "Användare", href: "/users", icon: "👥" },
];

export default function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const isActive = (href: string) => pathname?.startsWith(href);

  return (
    <>
      {/* Mobile menu button */}
      <div className="lg:hidden fixed top-0 left-0 right-0 z-50 bg-[#0B1514] border-b border-[rgba(255,255,255,0.06)]">
        <div className="flex items-center justify-between px-4 py-3">
          <img src="/eskio-logo.png" alt="Eskio" className="h-8" />
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="text-white p-3 hover:bg-[#142421] rounded-md text-2xl outline-none focus:outline-none"
          >
            {mobileMenuOpen ? "✕" : "☰"}
          </button>
        </div>
      </div>

      {/* Mobile menu overlay */}
      {mobileMenuOpen && (
        <div
          className="lg:hidden fixed inset-0 bg-black bg-opacity-50 z-40"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div className={`
        fixed lg:static inset-y-0 left-0 z-40
        flex h-full w-64 flex-col bg-[#0B1514]
        transform transition-transform duration-200 ease-in-out
        ${mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        lg:pt-0 pt-16
      `}>
      {/* Logo - hidden on mobile (shown in top bar instead) */}
      <div className="hidden lg:flex h-16 items-center px-6 bg-[#0F1D1A]">
        <img src="/eskio-logo.png" alt="Eskio" className="h-10" />
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3 py-4">
        {navigation.map((item) => (
          <Link
            key={item.name}
            href={item.href}
            onClick={() => setMobileMenuOpen(false)}
            className={`flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
              isActive(item.href)
                ? "bg-[var(--brand-light)] text-[var(--brand)]"
                : "text-gray-300 hover:bg-[#142421] hover:text-white"
            }`}
          >
            {item.icon === "ester" ? (
              <Image src="/ester-banner.png" alt="Ester AI" width={24} height={24} className="mr-3 rounded-full" />
            ) : (
              <span className="mr-3 text-lg">{item.icon}</span>
            )}
            {item.name}
          </Link>
        ))}

        {/* Admin section */}
        {user?.role === "Admin" && (
          <>
            <div className="pt-4 pb-2">
              <p className="px-3 text-xs font-semibold text-gray-400 uppercase tracking-wider">
                Administration
              </p>
            </div>
            {adminNavigation.map((item) => (
              <Link
                key={item.name}
                href={item.href}
                onClick={() => setMobileMenuOpen(false)}
                className={`flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                  isActive(item.href)
                    ? "bg-[var(--brand-light)] text-[var(--brand)]"
                    : "text-gray-300 hover:bg-[#142421] hover:text-white"
                }`}
              >
                <span className="mr-3 text-lg">{item.icon}</span>
                {item.name}
              </Link>
            ))}
          </>
        )}
      </nav>

      {/* Dark mode toggle */}
      <div className="px-3 pb-2">
        <button
          onClick={toggleTheme}
          className="flex items-center w-full px-3 py-2 text-sm font-medium rounded-md text-gray-300 hover:bg-[#142421] hover:text-white transition-colors"
        >
          <span className="mr-3 text-lg">{theme === "dark" ? "☀️" : "🌙"}</span>
          {theme === "dark" ? "Ljust läge" : "Mörkt läge"}
        </button>
      </div>

      {/* User section */}
      <div className="border-t border-[rgba(255,255,255,0.06)] p-4">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className="h-10 w-10 rounded-full bg-gray-700 flex items-center justify-center">
              <span className="text-lg">👤</span>
            </div>
          </div>
          <div className="ml-3 flex-1">
            <p className="text-sm font-medium text-white">{user?.name}</p>
            <p className="text-xs text-gray-400">{user?.role}</p>
          </div>
          <button
            onClick={() => logout()}
            className="ml-2 text-gray-400 hover:text-white transition-colors"
            title="Logga ut"
          >
            🚪
          </button>
        </div>
      </div>
    </div>
    </>
  );
}
