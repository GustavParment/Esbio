# Changelog

Notable changes grouped by session / branch. For day-to-day history, use `git log`.

---

## SIE import + Mini plan + SKV 4700 PDF + org.nr validation (2026-04-15)

**Status:** On `dev`, awaiting deploy.

### Momsdeklaration (SKV 4700) PDF export
- New endpoint `GET /api/v1/reports/vat/pdf?from_date=&to_date=`
- Generates a Skatteverket-styled PDF mapping existing VAT report data to boxes 05, 10/11/12, 42, 48, 49 — user transcribes to Skatteverket's e-tjänst
- "Ladda ner momsdeklaration (PDF)" button on `/reports/vat`

### SIE 4 import (Fortnox/Visma migration)
- New parser `service/sie_parser.go` — auto-detects encoding (UTF-8 / CP437 / ISO-8859-1), handles quoted strings with escapes, brace blocks, validates voucher balance
- New import service `service/sie_import_service.go` — preview (dry-run) + commit; infers BAS account type/side/group from account number; skips unbalanced vouchers
- New endpoints `POST /reports/sie/import/preview` and `POST /reports/sie/import` (multipart, 25MB cap)
- Frontend: `/inställningar` → "Importera bokföring (SIE)" section with file picker, preview card (company/period/account/voucher counts, warnings), confirm-then-import flow
- Test fixture at `server/testdata/sample_import.se`
- Parser unit tests in `service/sie_parser_test.go` (6 cases)
- **Bugfix:** initial implementation only persisted voucher headers; now also writes line items via `LineItemService`

### Mini plan (99 kr/mån)
- New plan tier added to Stripe price-to-plan mapping (`mini` for both test + live placeholders — replace `price_MINI_*_PLACEHOLDER` once Stripe products are created)
- `/upgrade` page: third Mini card with limited feature list
- Landing `/` pricing grid: Mini replaces Enterprise (which moves to small "Kontakta oss" line); pricing cards now use flexbox so all CTA buttons align horizontally
- Sidebar plan badge shows "Mini / 99 kr/mån"
- **Note:** no backend feature gating yet — Mini plan currently doesn't block AI/invoicing endpoints; gating middleware is a follow-up

### Org.nummer validation (mandatory)
- Shared FE helper `lib/utils/orgNumber.ts`: `formatOrgNumber()` auto-formats as XXXXXX-XXXX, `validateOrgNumber()` checks 10 digits + Luhn checksum
- Applied to `/companies/new` and `/inställningar` forms — auto-format on input, error on blur, blocks submit on invalid
- Backend defense in depth: `CompanyService.CreateCompany` and `UpdateCompany` reject empty/invalid org numbers and normalize to canonical form

### Mobile sidebar fix
- Logout + dark-mode toggle now appear inside nav (just below Hjälp) on mobile so iPhone Safari chrome doesn't hide them; desktop layout unchanged

---

## Email integration + auth improvements — DEPLOYED to prod (2026-04-14)

**Status:** Live on `https://esbio.se` and `https://api.esbio.se`. Backend revision `esbio-backend-00040-grn`, frontend revision `esbio-frontend-00018-kgv`. DB migration 014 applied to prod Cloud SQL. `RESEND_API_KEY` configured in Cloud Run.

### Invoice email (Fas 5)
- Send invoices via email with PDF attachment using Resend API
- "Skicka via e-post" button on invoice detail page (requires finalized invoice + customer email)
- Sender: `{CompanyName} <noreply@esbio.se>`, Swedish HTML template
- Resend domain verified with DKIM/SPF on Cloudflare

### Email verification & password reset (Fas 6)
- Verification email on registration (24h link), login blocked until verified
- Welcome email after successful verification
- Forgot password flow: request reset → email with 1h link → set new password
- "Glömt lösenord?" link on login page
- Frontend pages: `/auth/verify`, `/auth/forgot-password`, `/auth/reset-password`
- DB migration 014: `email_verified`, `verification_token`, `reset_token` columns on users

### OCR number improvement
- OCR now includes company ID + invoice number (10 digits + Luhn check), was previously just invoice number

### Support contact form
- Contact form in Settings page (subject + message)
- Sends email to info@esbio.se via Resend with reply-to set to user's email
- POST `/api/v1/support` endpoint

### Ester AI — invoice tools
- `list_invoices` — list invoices filtered by status
- `send_invoice_email` — send invoice via email to customer
- Ester can now handle "skicka alla obetalda fakturor" type commands

### Help page
- New `/help` page with 7 guide sections: Getting started, Bookkeeping basics, Ester AI, Receipt scanning, Invoicing, Reports, Bookkeeping tips
- Includes common journal entries, VAT rates, important tax dates, links to Skatteverket
- Added to sidebar navigation

### Settings cleanup
- Removed AI usage statistics section (confused users)

---

## On dev, awaiting testing — `feature/invoicing` (2026-04-12)

Invoicing module merged to `dev`. NOT on `main`/prod yet — awaiting manual testing.

### Invoicing module (Fas 1-4)

Full Swedish invoicing with auto-bokforing:
- **Customer registry**: CRUD + search (6 API endpoints, 3 frontend pages)
- **Invoice settings**: bankgiro, plusgiro, swish, F-skatt, payment terms, invoice number sequence
- **Invoice CRUD**: create draft, finalize (auto-generates sales voucher: 1510/3010/2610-2612/3740), mark as paid (payment voucher: bank/1510), cancel (correction voucher), delete draft
- **Invoice PDF**: Swedish layout via fpdf (company/customer info, lines, VAT summary, OCR, payment details, F-skatt)
- **Overdue detection**: scheduler checks all companies every 60s, flips sent→overdue past due date
- **Dashboard widget**: outstanding, overdue, and paid invoice summaries
- OCR number generation (Luhn), oresavrundning (round to whole krona, diff to 3740)
- 4 DB tables: `customers`, `invoice_settings`, `invoices`, `invoice_lines`
- 17 API endpoints, ~20 new Go files, ~10 new frontend pages

See [invoicing-plan.md](./invoicing-plan.md) and [invoicing-test-guide.md](./invoicing-test-guide.md).

### Mobile-first UI overhaul

Replaced all HTML tables across the entire app with card-based layouts:
- Lists: vouchers, invoices, customers, accounts, ledger, dashboard — all clickable cards
- Forms: voucher line editor, invoice line editor — vertical cards with full-width fields
- Detail views: voucher lines, invoice lines — compact card display
- Settings: full-width buttons, wrapping subscription section
- Sidebar: full screen width on mobile (was w-64 leaving black strip)
- Voucher detail: buttons wrap, Invalid Date bug fixed
- 12 files, net -200 lines

---

## Deployed to prod — `refactor/money-int64` (2026-04-11/12)

Deployed to prod on 2026-04-12. Stripe key rotated. All changes below are live.

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

**4. Stripe webhook — plan upgrade bypass chain** (`server/cmd/api/internal/handlers/stripe_handler.go`, `service/stripe_service.go`)

Found and verified a chain of three bugs in the Stripe webhook path that together allowed an unauthenticated `free → starter` plan upgrade via a single POST. Verified locally by running an exploit curl from the Kali pentest VM against the dev backend — the webhook returned `200 {"received":true}` and the DB row flipped to `plan=starter, status=active, stripe_subscription_id=sub_fake_exploit_001`. All three bugs now fixed and verified.

- **Signature fallback**: `stripe_handler.go:92-98` had an `if webhookSecret != "" { verify } else { parse anyway }` fallback. Prod currently has `STRIPE_WEBHOOK_SECRET` configured so the bypass wasn't actively exploitable there, but a single misconfigured deploy (missing env var, stripped secret, new region) would have re-opened it. Now fails closed: returns `503 {"error":"webhook not configured"}` with an ERROR-level log line when the secret is missing. No unverified webhook body is ever parsed.
- **`priceToPlan` default**: `stripe_service.go:199-201` returned `"starter"` for any unknown `price_id`, meaning a forged webhook with `"price_id":"price_anything"` would map to a paid tier. Now returns `""` and the caller rejects empty plans.
- **Nil pointer panic**: `stripe_service.go:137` dereferenced `sub.Items.Data` without a nil check, panicking on any subscription payload without items. sqlmap hit this during fuzzing and crashed the request. Added nil guards on `Items`, `Data[0]`, and `Price`.
- **Defense in depth**: `HandleSubscriptionEvent` now re-fetches the subscription directly from the Stripe API via `subscription.Get(sub.ID, nil)` and uses **only** that authoritative state. A forged `sub_fake_xxx` ID will 404 at Stripe before any DB update happens — even if signature verification were somehow bypassed, the forged event can't drive an upgrade.

Post-fix verification from the same Kali VM: exploit curl returns `503`, DB row unchanged. See [pentest-guide.md §"What was found during the 2026-04-11 session"](./pentest-guide.md#what-was-found-during-the-2026-04-11-session) for full regression steps.

### New documentation

- [pentest-guide.md](./pentest-guide.md) — internal pentest playbook covering scope, environment setup, 10+ Kali Linux tools (sqlmap, ffuf, nuclei, jwt_tool, Burp, semgrep, gitleaks, testssl.sh, etc.), Esbio-specific business-logic tests (company isolation, plan upgrade, mass assignment, LLM tool injection), and cleanup procedures. Includes the exact sqlmap commands used in this session and the full exploit trace for the Stripe bugs.

### Operational state changes

- **Prod backend paused** via `gcloud run services update esbio-backend --ingress=internal` on 2026-04-11. External requests to `api.esbio.se` are rejected at Google's edge. Unpause with `--ingress=all`.
- **Cloud SQL `esbio-db`** still has public IP (`34.88.55.83`) with whatever authorized networks are configured. Not yet migrated to private IP. See [network-access-setup.md](./network-access-setup.md) and the TODO below.
- **`STRIPE_SECRET_KEY` (live) must be rotated**: during the pentest session the live key was echoed to the terminal via `gcloud run services describe`. It's now in shell scrollback and may persist in logs. Rotate in Stripe Dashboard → Developers → API keys → Roll, then update the Cloud Run env var.

### Outstanding TODOs

- Browser click-through of the refactored frontend against local dev backend
- **Rotate `STRIPE_SECRET_KEY`** (live key leaked to terminal during pentest session)
- Redeploy `main` to Cloud Run and unpause prod (`--ingress=all`)
- Migrate `esbio-db` to private IP + Cloud SQL Auth Proxy (still planned; original proposal was option B — `gcloud run services update --add-cloudsql-instances=...`)
