"use client";

import Link from "next/link";

interface ButtonProps {
  children: React.ReactNode;
  variant?: "primary" | "secondary" | "outline" | "danger" | "ghost";
  size?: "sm" | "md" | "lg";
  href?: string;
  onClick?: (e: React.MouseEvent) => void;
  disabled?: boolean;
  type?: "button" | "submit";
  className?: string;
  fullWidth?: boolean;
}

const variants = {
  primary: "bg-[var(--brand)] text-white hover:bg-[var(--brand-hover)] disabled:opacity-50",
  secondary: "bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700",
  outline: "border border-gray-300 text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-200 dark:hover:bg-gray-800",
  danger: "bg-red-600 text-white hover:bg-red-700 disabled:opacity-50",
  ghost: "text-[var(--brand)] hover:bg-[var(--brand-light)]",
};

const sizes = {
  sm: "px-3 py-1.5 text-xs rounded-lg",
  md: "px-4 py-2 text-sm rounded-lg",
  lg: "px-6 py-3 text-base rounded-xl",
};

export default function Button({
  children,
  variant = "primary",
  size = "md",
  href,
  onClick,
  disabled = false,
  type = "button",
  className = "",
  fullWidth = false,
}: ButtonProps) {
  const classes = `
    inline-flex items-center justify-center font-medium transition-colors
    disabled:cursor-not-allowed
    ${variants[variant]}
    ${sizes[size]}
    ${fullWidth ? "w-full" : ""}
    ${className}
  `.trim();

  if (href) {
    return (
      <Link href={href} className={classes}>
        {children}
      </Link>
    );
  }

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={classes}
    >
      {children}
    </button>
  );
}
