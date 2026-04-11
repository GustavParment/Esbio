# Money Refactor Plan — replace `float64` with `decimal.Decimal`

**Branch:** `refactor/money-int64` (name kept for history; actual approach uses `decimal.Decimal`, not `int64` öre — see §2)
**Status:** ✅ Implemented (uncommitted — 2026-04-11), pending production deploy
**Author:** Gustav + Claude
**Date:** 2026-04-11

---

## 0. Status summary (2026-04-11)

All six phases complete. Backend (`go build`, `go vet`) and frontend (`npx tsc --noEmit`) are clean. Live end-to-end test on the local dev DB confirmed:

- **0.1 + 0.2 == 0.3** now returns `balanced: true` from the voucher validation endpoint (would have failed the old float64 check)
- Values round-trip exactly through `PG DECIMAL(15,2) → Go decimal.Decimal → JSON strings`
- Income statement, balance sheet, and VAT report aggregations all return exact decimal strings
- The LLM tool-call path (`create_voucher`) successfully creates a voucher with VAT calculation (`50184 × 1.25 = 62730` exact)

**Not yet done:** commit, merge `refactor/money-int64 → dev → main`, redeploy Cloud Run, unpause prod ingress.

---

## 1. Problem statement

Every monetary field in the Go backend is typed as `float64`. For a bookkeeping
system this is a correctness bug:

- IEEE 754 binary floats cannot represent decimal fractions exactly. `0.1 + 0.2`
  evaluates to `0.30000000000000004`.
- Across many rows (ledger sums, VAT totals, balance sheets) the drift
  accumulates into visible öre-level discrepancies.
- Debit/credit equality checks can fail even when the underlying voucher is
  balanced.
- SIE exports and VAT reports sent to Skatteverket may be off by öre and fail
  validation.
- Any downstream system that reconciles against bank data will see phantom
  mismatches.

### Good news: the database is not the problem

All monetary columns in `server/cmd/api/migrations/*.sql` and `server/init.sql`
are `DECIMAL(15, 2)`:

```
line_items.debit_amount    DECIMAL(15, 2)
line_items.credit_amount   DECIMAL(15, 2)
vouchers.total_amount      DECIMAL(15, 2)
bank_accounts.balance_amount DECIMAL(15, 2)
bank_transactions.amount   DECIMAL(15, 2)
```

PostgreSQL stores and computes these exactly. The precision loss happens
**only when Go's `lib/pq` driver scans `DECIMAL` into `float64`** in the
repository layer, and any Go-side arithmetic afterwards.

**Implication: no schema migration, no data backfill, no × 100 rescaling.**
The existing rows in prod are already correct; we just need to stop
round-tripping them through `float64`.

---

## 2. Chosen approach: `github.com/shopspring/decimal`

Two real options were considered:

| Option | Pros | Cons |
|---|---|---|
| **A. `int64` öre** | Fast, no deps, zero-allocation math | Requires × 100 conversions at every DB boundary; requires DB schema migration to `BIGINT`; every JSON contract changes; frontend must know "öre vs kronor" |
| **B. `decimal.Decimal`** | Scans directly from `DECIMAL` via `database/sql.Scanner`; exact arbitrary-precision math; no DB migration; natural decimal semantics; well-maintained library | Slightly slower than int64; allocations on arithmetic; JSON marshals as string by default (which is actually correct — see §6) |

**Decision: Option B (`decimal.Decimal`).** With the DB already on `DECIMAL(15,2)`,
Option A would be work for no structural benefit — we'd be introducing a second
representation (öre) that doesn't match the storage. `shopspring/decimal` is the
idiomatic Go choice and integrates cleanly with `database/sql`.

Package: `github.com/shopspring/decimal`
License: MIT
Used by: Uber, Stripe Go clients, many financial projects

---

## 3. Scope — files and symbols touched

### 3.1 Domain models (`server/cmd/api/internal/domain/models.go`)

Change `float64` → `decimal.Decimal` for:

- `Account.Balance`
- `LineItem.DebitAmount`, `CreditAmount`
- `Voucher.TotalAmount`
- `LedgerEntry.DebitAmount`, `CreditAmount`, `Balance`
- `IncomeStatementEntry.Balance`
- `IncomeStatement.TotalIncome`, `TotalExpenses`, `NetResult`
- `BalanceSheetEntry.Balance`
- `BalanceSheet.TotalAssets`, `TotalEquityLiab`, `NetResult`
- `VATReportEntry.TotalSales`, `TotalVAT`
- `VATReport.TotalSales`, `TotalVAT`, `TotalInputVAT`, `NetVAT`
- `BankAccount.BalanceAmount` (pointer — `*decimal.Decimal`)
- `BankTransaction.Amount`
- `ReceiptScan.TotalAmount`, `VATAmount`, `AmountExclVAT`
- `CorrectionLine.DebitAmount`, `CreditAmount`
- `AgentUsageSummary.EstimatedCostSEK`, `MonthEstimatedCostSEK` (see §3.6 — exception)

**Explicitly NOT changed (not money):**
- `ReceiptScan.Confidence` — a 0.0–1.0 probability, stays `float64`
- LLM token cost calculations inside `agent_service.go` and
  `agent_usage_repository.go` — these are estimates over USD pricing, not
  accounting values. They can stay `float64` and only convert to `decimal` at
  the final display boundary if shown to the user in SEK.

### 3.2 Repositories (`server/cmd/api/internal/repository/`)

- `line_item_repository.go` — scan into `decimal.Decimal`, pass in queries
- `voucher_repository.go` — `total_amount`
- `account_repository.go` — ledger queries, running balance computation
- `report_repository.go` — P&L/BS sums, VAT aggregation (currently has its
  own `float64` accumulators we need to replace with `decimal.Sum`)
- `bank_account_repository.go`, `bank_transaction_repository.go`

`shopspring/decimal` implements `sql.Scanner` and `driver.Valuer`, so
`rows.Scan(&lineItem.DebitAmount)` works directly against a `DECIMAL` column.

### 3.3 Services (`server/cmd/api/internal/service/`)

- `voucher_service.go` — debit/credit totaling, balance validation
  (`totalDebit.Equal(totalCredit)` instead of `==`)
- `report_service.go` — any aggregation
- `sie_generator.go` — IB/UB maps, period movement; SIE file format uses
  decimal strings with dot separator, use `.String()` or a custom formatter
- `agent_service.go` — tool arg parsing (currently does `float64` type
  assertions on JSON-decoded args; need to accept string or number and
  convert to `decimal.Decimal`)

### 3.4 Handlers (`server/cmd/api/internal/handlers/`)

- `voucher_handler.go` — request body structs use `float64`, change to
  `decimal.Decimal` (which unmarshals from JSON string OR number)
- `pdf_handler.go` — `formatCurrency(amount float64) string` becomes
  `formatCurrency(amount decimal.Decimal) string`
- `report_handler.go`, `receipt_handler.go` — wherever money flows in/out

### 3.5 JSON API contract

`decimal.Decimal` marshals to a **JSON string** by default (e.g. `"1234.56"`).
This is a breaking change if the frontend currently expects numbers.

**Options:**
1. **Keep it as string** (recommended). JavaScript's `number` cannot safely
   represent arbitrary decimals either — you have the same bug client-side
   waiting to happen. Use strings and either format for display or use
   `decimal.js` in the frontend for any calculations.
2. Use `decimal.NewFromFloat` in a custom `MarshalJSON` that emits numbers —
   but that just re-introduces the float bug at the wire level. Don't.

We'll go with **strings** and update frontend types accordingly (see §4).

### 3.6 Exception: agent cost tracking

`agent_usage_repository.go` computes LLM costs like:
```go
usdCost := float64(promptTokens)*0.15/1000000 + float64(completionTokens)*0.60/1000000
```

This is a display-only estimate, not accounting data. It's **not stored in
`DECIMAL`** columns, it's not part of the ledger, and Skatteverket will never
see it. Leave as `float64` to avoid dragging the refactor into non-financial
code.

---

## 4. Frontend changes (`client/`)

All TypeScript interfaces that describe money-bearing API responses need to
change `number` → `string`. Then:

- **Display**: `parseFloat(amount).toLocaleString('sv-SE', { minimumFractionDigits: 2, maximumFractionDigits: 2 })` is acceptable for one-shot rendering. The float error is bounded and invisible at 2 decimals.
- **Calculations**: if the frontend ever sums or multiplies money (e.g. live
  voucher preview), use `decimal.js` to mirror the backend.
- **Form inputs**: keep `<input type="number">` but convert to string before
  sending to the API.

Files to grep for (not exhaustive — do this as part of step 6):
```
grep -rn "debit_amount\|credit_amount\|total_amount\|balance" client/src --include="*.ts" --include="*.tsx"
```

---

## 5. Database migration

**None required.** Columns are already `DECIMAL(15, 2)`.

Verification step (should return 0 rows):
```sql
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'public'
  AND column_name IN ('debit_amount', 'credit_amount', 'total_amount', 'balance', 'amount', 'balance_amount')
  AND data_type NOT IN ('numeric');
```

---

## 6. Implementation order

Do this in small PRs against the `refactor/money-int64` branch, each
independently compilable:

### Phase 1 — foundation
1. `go get github.com/shopspring/decimal`
2. Change `domain/models.go` types
3. Fix compile errors file-by-file in the order: repository → service →
   handler. Each file must compile green before moving on.

### Phase 2 — arithmetic correctness
4. Replace all `==`, `!=`, `+`, `-`, `*`, `/` on money with
   `.Equal`, `.Add`, `.Sub`, `.Mul`, `.Div` calls
5. Replace `float64` accumulators (`var total float64`) with
   `decimal.Zero` and `total = total.Add(row.Amount)`
6. Voucher balance check: `totalDebit.Equal(totalCredit)` with no epsilon

### Phase 3 — SIE and PDF formatting
7. Update `sie_generator.go` — SIE format is strict, use `.StringFixed(2)`
   with locale dot separator
8. Update `pdf_handler.go` — `formatCurrency` reads a `decimal.Decimal`

### Phase 4 — agent tool args
9. `agent_service.go` currently does `args["debit_amount"].(float64)` when
   the LLM passes tool call arguments. JSON numbers still decode as
   `float64`, so: accept the `float64`, then `decimal.NewFromFloat(v)` with
   a clear understanding this is the **one place** float comes in. The LLM's
   tool args are not authoritative accounting data — the user confirms the
   voucher in the UI anyway.

### Phase 5 — frontend
10. Update TypeScript types: `number` → `string` for money fields
11. Fix all call sites (display + any calculation)
12. Add `decimal.js` if any client-side math exists

### Phase 6 — testing
13. Unit tests for voucher balance, sums, VAT calculation — use known
    pathological inputs (many `0.1 + 0.2` style sums)
14. End-to-end test: create a voucher with 100 lines of 0.01 SEK, verify
    total is exactly 1.00 SEK
15. Generate a SIE file against the dev DB, diff against previous output —
    should differ only in öre-level corrections, not in format

### Phase 7 — deploy
16. Merge `refactor/money-int64` → `dev`, deploy dev, smoke test manually
17. Verify a real voucher round-trip: create → read → PDF → SIE
18. Merge `dev` → `main`, deploy prod (off-hours)
19. Monitor logs and error rates for 24h

---

## 7. Risk register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Frontend breaks because JSON fields are now strings | High | Medium | Update types in lockstep; don't deploy backend alone |
| LLM tool calls pass malformed numbers | Medium | Low | `decimal.NewFromString` with error handling; fall back to float for agent layer |
| Existing vouchers fail to load due to scanner mismatch | Low | High | `shopspring/decimal` scans from any numeric type; test against a copy of prod DB first |
| Subtle drift in historical reports post-refactor | Low | Low | Reports will become **more** accurate, not less; any diff is a fix |
| SIE format regression | Medium | High | Keep a golden-file test: run SIE export on known fixture, diff |
| Increased allocation pressure in hot paths | Low | Low | `decimal.Decimal` is immutable — allocations happen but are small; profile if it becomes an issue |

---

## 8. Out of scope (explicitly)

- DB schema changes (not needed)
- Migrating existing data (not needed)
- Changing `Confidence` (not money)
- Changing LLM cost tracking (not accounting data)
- Adding new currencies (single-currency SEK remains)
- Performance optimization of decimal math (only if profiling shows it)

---

## 9. Estimated effort

- Backend: 1–1.5 days (touches ~25 files but mechanically)
- Frontend: 0.5 day (types + display call sites)
- Tests: 0.5 day (unit + e2e)
- Review + deploy: 0.5 day
- **Total: ~3 days of focused work**

---

## 10. Decisions (resolved 2026-04-11)

1. **JSON format**: **strings**. `decimal.Decimal` will marshal as
   `"1234.56"`. Frontend types change `number` → `string`. Rationale:
   JSON numbers decode into IEEE 754 floats in every language, reintroducing
   the exact bug we're fixing. Strings pass through untouched and match the
   convention used by every serious financial API (PayPal, Plaid, Adyen,
   OpenBanking PSD2).
2. **Staging strategy**: no dedicated staging env; validate on the
   `refactor/money-int64` branch against a **copy of prod DB** restored
   locally before any deploy. Dump via `gcloud sql export sql` into a GCS
   bucket, download, restore into local Postgres.
3. **Client-side math**: unknown — audit pending as part of Phase 5 kickoff.
   Grep command: `grep -rn "debit_amount\|credit_amount\|total_amount\|balance\|\.amount" client/src --include="*.ts" --include="*.tsx"`.
4. **Pre-deploy testing**: mandatory. No code from this branch deploys to
   prod (or even `dev` cloud) until it has been validated end-to-end against
   a copy of real data on the refactor branch.
