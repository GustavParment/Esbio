# Changelog

Notable changes grouped by session / branch. For day-to-day history, use `git log`.

---

## Unreleased — branch `refactor/money-int64` (2026-04-11)

All changes below are **uncommitted** on the `refactor/money-int64` branch and not yet deployed. Prod backend is still paused (`ingress=internal`, applied 2026-04-11). See individual doc links for full details.

### Fixes

**1. Gin trusted-proxy configuration** (`server/cmd/api/main.go`)

`gin.Default()` was running with the default "trust every proxy" setting, emitting a `[WARNING] You trusted all proxies, this is NOT safe` line on every cold start and causing `c.ClientIP()` to potentially honor attacker-supplied `X-Forwarded-For` headers. Added explicit configuration:

```go
router.ForwardedByClientIP = true
router.SetTrustedProxies([]string{"0.0.0.0/0"})
```

The `0.0.0.0/0` is intentional and safe because Cloud Run is the only path into the container — nothing can bypass Google's front end to hit Gin directly. Rate limiting and access logs now see real client IPs instead of the LB IP.

**2. Money precision — `float64` → `decimal.Decimal`** (backend + frontend)

Bookkeeping data was carried through Go as `float64` despite the DB already using `DECIMAL(15, 2)`. IEEE 754 floats can't exactly represent decimal fractions (`0.1 + 0.2 = 0.30000000000000004`), causing subtle öre-level drift in ledger sums, VAT totals, and balance checks.

Replaced with `github.com/shopspring/decimal` end-to-end:

- 20+ fields in `domain/models.go` flipped from `float64` to `decimal.Decimal`
- All repositories (`account_repository.go`, `report_repository.go`, etc.) — scan directly into `decimal.Decimal`, accumulators use `.Add`/`.Sub`
- Services — `voucher_service.go` balance check is now exact (`totalDebit.Equal(totalCredit)`), `report_repository.go` VAT calc uses `.Mul`/`.Div`, `sie_generator.go` emits `.StringFixed(2)`
- Handlers — `voucher_handler.go` request body, `pdf_handler.go` `formatCurrency`, etc.
- Agent — new `decimalFromArg()` helper isolates the LLM tool-call entry point (the one place where float64 enters, bounded and UI-confirmed)

Frontend:
- New `client/lib/money.ts` with `parseMoney`, `formatSEK`, `sumMoney`, `isPositive`, `isZero`
- `client/types/index.ts` — new `MoneyString` type alias, all money fields flipped from `number` to `MoneyString`
- `client/lib/api/reports.ts` — same treatment
- 9 page components updated: dashboard, vouchers list/detail/scan, account ledger, and all 3 reports (balance-sheet, income-statement, vat)

JSON contract change: money fields now travel as strings (`"1234.56"`) instead of numbers. Requests accept either, responses always return strings. No DB migration was needed — schema was already correct.

**Validated live**: posted a voucher with `0.1 + 0.2 = 0.3` lines, `/vouchers/:id/validate` returned `balanced: true`. Would have failed under the old float code. See [money-refactor-plan.md](./money-refactor-plan.md) for the full plan, phases, and file list.

**3. Ester AI streaming — empty-response fallback + model bump** (`server/cmd/api/internal/service/agent_service.go`)

Observed: user creates a voucher via `/agent/stream`, the tool executes and succeeds, but the SSE stream ends with `done` without emitting any `text` events. User sees a spinning indicator, no confirmation message.

Root cause: `gemini-2.5-flash-lite` often returns an empty final turn after a successful tool call (the model assumes the tool output is self-explanatory). The streaming loop entered the `!hasFunctionCall` branch with an empty `textResult`, called `streamText("")` (no-op), and closed the stream.

Fixes:
- Both the blocking `Chat` and the streaming `ChatStream` now call full `gemini-2.5-flash` (not `-lite`), which reliably produces final-turn summaries
- Added a `lastToolResult` tracker — if the LLM's final turn has no text, the tool's own human-readable return (`"Verifikat #1 skapat (ID: 83) med 3 konteringsrader..."`) is streamed instead
- Same fallback applies if the 10-iteration tool loop hits its cap (was silent before)
- Empty-candidates branch now also streams a user-visible message
- `streamText` chunk size `3 → 8` chars and delay `15ms → 5ms` — ~8× faster perceived typing, early-return on empty
- Receipt OCR (`receipt_service.go`) still uses `flash-lite` — single-shot vision, no tool calls, lite is fine there

See [ester-ai.md](./ester-ai.md#streaming-post-agentstream) for the streaming contract.

### Operational state changes

- **Prod backend paused** via `gcloud run services update esbio-backend --ingress=internal` on 2026-04-11. External requests to `api.esbio.se` are rejected at Google's edge. Unpause with `--ingress=all`.
- **Cloud SQL `esbio-db`** still has public IP (`34.88.55.83`) with whatever authorized networks are configured. Not yet migrated to private IP. See [network-access-setup.md](./network-access-setup.md) and the TODO below.

### Outstanding TODOs

- Browser click-through of the refactored frontend against local dev backend
- Split the agent streaming fix onto `dev` as its own PR (or keep bundled in the money refactor — decision pending)
- Commit everything on `refactor/money-int64`, merge `→ dev → main`, redeploy, unpause prod
- Migrate `esbio-db` to private IP + Cloud SQL Auth Proxy (still planned; original proposal was option B — `gcloud run services update --add-cloudsql-instances=...`)
