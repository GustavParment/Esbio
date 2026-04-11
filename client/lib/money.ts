// Money values arrive from the backend as decimal strings ("1234.56") to
// preserve precision across the wire. These helpers wrap parsing and
// display formatting so the application code doesn't touch raw strings.
//
// parseMoney converts to a JS number for display arithmetic only. Any
// authoritative math (voucher totals, balances, reports) must happen on
// the backend where decimal.Decimal is exact. The small rounding error
// JS introduces at the UI layer is bounded and acceptable for rendering.

import type { MoneyString } from "@/types";

export type MoneyInput = MoneyString | number | null | undefined;

export function parseMoney(value: MoneyInput): number {
  if (value === null || value === undefined || value === "") return 0;
  if (typeof value === "number") return value;
  const n = parseFloat(value);
  return Number.isFinite(n) ? n : 0;
}

export function formatSEK(value: MoneyInput): string {
  return parseMoney(value).toLocaleString("sv-SE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function isPositive(value: MoneyInput): boolean {
  return parseMoney(value) > 0;
}

export function isZero(value: MoneyInput): boolean {
  return parseMoney(value) === 0;
}

// Sum a list of money-bearing items. Returns a JS number suitable for
// display. For authoritative totals, read the value from the backend.
export function sumMoney<T>(items: T[], selector: (item: T) => MoneyInput): number {
  let total = 0;
  for (const item of items) {
    total += parseMoney(selector(item));
  }
  return total;
}
