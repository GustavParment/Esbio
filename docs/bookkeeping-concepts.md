# Bookkeeping Concepts

This document explains the accounting concepts implemented in Esbio for developers who may not be familiar with Swedish bookkeeping.

## Double-Entry Bookkeeping

Every financial transaction is recorded as a **voucher** (verifikat) with at least two **line items**. The fundamental rule: **total debits must equal total credits**.

Example — recording a sale of 10,000 kr:

| Account       | Debit     | Credit    |
|---------------|-----------|-----------|
| 1930 Bankkonto | 10,000   |           |
| 3010 Forsaljning |         | 10,000   |

## BAS Chart of Accounts (Kontoplan)

Esbio uses the Swedish **BAS** standard for organizing accounts into 8 groups:

| Group | Name (Swedish)              | Name (English)           | Type | Normal Side |
|-------|-----------------------------|--------------------------|------|-------------|
| 1     | Tillgangar                  | Assets                   | BS   | Debit       |
| 2     | Eget kapital & skulder      | Equity & Liabilities     | BS   | Credit      |
| 3     | Intakter                    | Revenue / Income         | P&L  | Credit      |
| 4     | Material & varor            | Cost of Goods            | P&L  | Debit       |
| 5     | Ovriga externa kostnader    | Other External Costs     | P&L  | Debit       |
| 6     | Personal                    | Personnel Costs          | P&L  | Debit       |
| 7     | Avskrivningar               | Depreciation             | P&L  | Debit       |
| 8     | Finansiella poster          | Financial Items          | P&L  | Debit       |

**Type:**
- **BS** (Balance Sheet) — accounts that track cumulative balances (assets, liabilities, equity)
- **P&L** (Profit & Loss) — accounts that track income and expenses for a period

**Normal Side:** The side (Debit/Credit) where increases are recorded.

## Key Terminology

| Swedish         | English              | Description                                        |
|-----------------|----------------------|----------------------------------------------------|
| Verifikat       | Voucher              | A journal entry recording a transaction            |
| Kontoplan       | Chart of Accounts    | The list of all accounts                           |
| Huvudbok        | General Ledger       | Transaction history per account                    |
| Resultatrakning | Income Statement     | Profit & loss report for a period                  |
| Period          | Period               | Accounting month (YYYY-MM format)                  |
| Debet           | Debit                | Left side of an entry                              |
| Kredit          | Credit               | Right side of an entry                             |
| Momskod         | Tax Code (VAT)       | Swedish VAT rate: 25%, 12%, 6%, or 0%             |

## Vouchers (Verifikat)

A voucher represents one accounting event and contains:
- **Date** — when the transaction occurred
- **Description** — what the transaction is for
- **Reference** — invoice number, receipt ID, etc.
- **Period** — the accounting month (YYYY-MM)
- **Line items** — the debit/credit entries

Each voucher gets an auto-incrementing **voucher number** (#1, #2, #3...).

### Voucher Corrections

Instead of editing or deleting vouchers (which would break the audit trail), Esbio supports **corrections**:

1. **Simple reversal** — creates a new voucher with debits and credits swapped, effectively zeroing out the original
2. **Correction with changes** — creates a new voucher with the corrected values

The original voucher is marked as corrected and excluded from reports and ledger views.

## Account Ledger (Huvudbok)

The ledger shows all transactions for a specific account in chronological order with a **running balance**. Corrected vouchers are excluded to show only the effective transaction history.

## Income Statement (Resultatrakning)

A report that summarizes income and expenses for a date range:
- **Income** — sum of all group 3 account transactions (3000-3999)
- **Expenses** — sum of all group 4-8 account transactions (4000-8999)
- **Net result** — income minus expenses

Corrected vouchers are excluded from the calculation.

## Swedish VAT (Moms)

Line items can have a tax code representing the Swedish VAT rate:
- **25%** — standard rate (most goods and services)
- **12%** — food, restaurant services, hotel
- **6%** — books, newspapers, public transport
- **0%** — exempt (financial services, healthcare, education)
