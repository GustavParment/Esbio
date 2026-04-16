# Architecture Overview

Esbio is a full-stack Swedish bookkeeping application built around double-entry accounting principles using the BAS (Bas Kontoplan) chart of accounts.

## Tech Stack

| Layer        | Technology                          |
|--------------|-------------------------------------|
| Frontend     | Next.js 16, React 19, TypeScript    |
| Styling      | Tailwind CSS 4                      |
| Backend      | Go with Gin web framework           |
| Database     | PostgreSQL 15 — Cloud SQL (prod), Docker (dev) |
| Auth         | JWT (HS256) in httpOnly cookies     |
| PDF          | go-pdf/fpdf                         |
| Email        | Resend API (noreply@esbio.se)       |
| Payments     | Stripe Checkout + webhooks          |

## High-Level Architecture

```
┌─────────────────────┐       ┌──────────────────────────────────┐       ┌────────────┐
│   Next.js Client    │──────▶│        Go API Server             │──────▶│ PostgreSQL │
│   :3000             │ HTTP  │        :8080                     │  SQL  │   :5433    │
│                     │◀──────│                                  │◀──────│            │
└─────────────────────┘       └──────────────────────────────────┘       └────────────┘
```

## Backend Layers

The server follows a clean layered architecture:

```
HTTP Request
    │
    ▼
┌──────────────┐
│  Middleware   │  Auth (JWT), CORS, Role-based access, CompanyMiddleware
└──────┬───────┘
       ▼
┌──────────────┐
│  Handlers    │  Parse HTTP requests, validate input, return responses
└──────┬───────┘
       ▼
┌──────────────┐
│  Services    │  Business logic, validation rules, orchestration
└──────┬───────┘
       ▼
┌──────────────┐
│ Repositories │  SQL queries, database access
└──────┬───────┘
       ▼
   PostgreSQL
```

### Dependency Injection

All layers are wired up in `server/cmd/api/main.go`:

1. Load config from `.env`
2. Connect to PostgreSQL
3. Create JWT manager
4. Instantiate repositories (User, Company, Account, LineItem, Voucher, Report)
5. Instantiate services with their repositories (incl. `SIEGenerator` for export and `SIEImportService` for import — the latter wires `VoucherService`, `AccountService`, and `LineItemService` together because creating a voucher and its line items are two separate writes)
6. Instantiate handlers with their services
7. Register routes with Gin router
8. Start server

## Frontend Structure

The client uses the **Next.js App Router** with client-side rendering for protected routes.

```
app/
├── layout.tsx              Root layout with AuthProvider
├── page.tsx                Redirects to /auth/login
├── auth/                   Login, register, verify email, forgot/reset password
├── companies/              Company selector ("Mina Företag") + create company
├── dashboard/              Main dashboard
├── vouchers/               Voucher CRUD + corrections
├── accounts/               Account CRUD + ledger
├── reports/                Income statement, balance sheet, VAT report
├── agent/                  AI assistant chat + scheduled tasks
├── settings/               Company settings, support contact form
├── help/                   User guides (bookkeeping, Ester, invoicing, etc.)
├── invoices/               Invoice CRUD, PDF, send email
├── customers/              Customer registry
└── users/                  User management (Admin)

components/
├── layout/                 DashboardLayout, Sidebar, ProtectedRoute
└── ui/                     Reusable components (AccountSearch)

lib/
├── api/                    Typed API client modules
└── contexts/               AuthContext (React Context)

types/
└── index.ts                All TypeScript interfaces
```

### State Management

- **Global state:** React Context for authentication (`AuthContext`) and company selection
- **Local state:** `useState` / `useEffect` in page components
- **No external state library** (no Redux, Zustand, etc.)

### API Client

A base `ApiClient` class (`lib/api/client.ts`) handles:
- JSON serialization
- `credentials: 'include'` for cookie-based auth
- Error handling with typed `ApiError`
- Generic response typing

Domain-specific modules (`auth.ts`, `vouchers.ts`, `accounts.ts`, etc.) expose typed methods that call the base client.

## Authentication & Company Selection Flow

```
1. User submits credentials at /auth/login
2. Server validates, creates JWT (7-day expiry)
3. Server sets token in httpOnly cookie
4. User is redirected to /companies (company selector)
5. User selects or creates a company
6. Server sets company_id in a cookie via POST /companies/select
7. User is redirected to /dashboard
8. All subsequent API calls include both the auth and company_id cookies
9. AuthMiddleware validates the JWT on protected endpoints
10. CompanyMiddleware verifies the user owns the selected company
11. On app load, AuthContext calls /auth/me to restore session
12. ProtectedRoute requires both authentication AND a selected company
```

### Multi-Company Architecture

Users can have multiple companies. All data (vouchers, reports, AI conversations, scheduled tasks) is scoped per company, not per user. The selected company is tracked via a `company_id` cookie set when the user selects a company at `/companies`. The `CompanyMiddleware` verifies that the authenticated user owns the selected company before allowing access.

## Ester AI (Bookkeeping Assistant)

Esbio includes **Ester AI**, an intelligent bookkeeping assistant powered by Gemini 2.5 Flash (Google).

```
User message → Agent Handler → Gemini API (with tool definitions)
                                      │
                              Gemini decides to call tools
                                      │
                              Agent executes tools against existing services
                              (company_id enforced server-side)
                                      │
                              Result returned to Gemini → response to user
```

**Key components:**
- **AgentService** — orchestrates Gemini API calls with a tool-use loop (max 10 iterations)
- **14 tools** — wrapping existing services (voucher CRUD incl. corrections and search, reports, account ledger, scheduled tasks, invoice listing and email sending)
- **SchedulerService** — background goroutine checking for due tasks every 60 seconds
- **ScheduledTaskService** — CRUD for recurring monthly tasks with next-run calculation
- **System prompt** — Swedish-language, BAS-aware, with security rules against prompt injection
- **Error sanitization** — database errors are never exposed to users; sanitized messages are returned instead

**Security:** Company ID is enforced server-side on all tool calls. All data is scoped to the selected company. See [ester-ai.md](./ester-ai.md) for full security documentation.

## Security

For the internal pentest playbook (tools, commands, Esbio-specific tests, findings), see [pentest-guide.md](./pentest-guide.md).

- Passwords hashed with **bcrypt**
- **Email verification** required before login (24h token, blocks unverified users)
- **Password reset** via email with 1h expiring token (response always 200 to prevent enumeration)
- JWT stored in **httpOnly cookies** (not accessible via JavaScript)
- CORS whitelist for allowed origins
- **CompanyMiddleware** verifies company ownership on every request
- All data queries filter by `company_id` for strict data isolation between companies
- Role-based middleware (`RequireRole("Admin")`) for sensitive endpoints
- Input validation via `go-playground/validator`
- Foreign key constraints with `ON DELETE RESTRICT` / `CASCADE`
- **Gin trusted proxies** are explicitly configured (`router.SetTrustedProxies(["0.0.0.0/0"])` + `ForwardedByClientIP = true`) so `c.ClientIP()` returns the real user IP from `X-Forwarded-For` set by Google Cloud Run's front end. Safe because Cloud Run is the only ingress path to the container.

## Money representation

Monetary values are handled with `github.com/shopspring/decimal` end-to-end:

- **Database**: `DECIMAL(15, 2)` columns (always exact)
- **Go**: `decimal.Decimal` fields in all domain models, repositories, services, handlers — arithmetic uses `.Add`, `.Sub`, `.Mul`, `.Div`, comparisons use `.Equal` / `.GreaterThan` / `.IsZero`, never `==` or `+`/`-`
- **JSON on the wire**: decimal strings (`"1234.56"`) — JSON numbers decode to IEEE 754 floats in every language, reintroducing precision loss, so strings are the only safe wire format
- **Frontend (TypeScript)**: money-bearing fields are typed as `MoneyString` (an alias for `string`). Use `parseMoney()` and `formatSEK()` from `client/lib/money.ts` at the display boundary; do not do client-side arithmetic on money without a decimal library
- **Exception**: LLM token cost tracking in `agent_usage_repository.go` stays on `float64` — it's a display-only estimate, not ledger data

This is an end-to-end correctness guarantee: `0.1 + 0.2` equals `0.3` exactly, voucher balance checks use strict equality (no epsilon), and SIE4 exports emit `.StringFixed(2)` directly from decimal.

See [money-refactor-plan.md](./money-refactor-plan.md) for the rationale and rollout history.

## SIE 4 Import

SIE (Standard Import Export) is the Swedish standard for exchanging bookkeeping data. Esbio supports importing SIE 4 files to onboard existing companies.

**Files:** `sie_parser.go`, `sie_import_service.go`, `sie_import_handler.go`

**Endpoints:**
- `POST /reports/sie/import/preview` — dry-run, returns summary without persisting
- `POST /reports/sie/import` — parses, validates, and persists

**Encoding auto-detection** (three-tier):
1. Valid UTF-8 containing Swedish characters (åäöÅÄÖ) → UTF-8
2. `#FORMAT PC8` directive in first 1KB → CP437
3. Fallback → ISO-8859-1

**Preview vs Commit:**
- `PreviewImport()` parses the file and returns an `SIEImportSummary` (company name, org number, fiscal dates, encoding, account/voucher/unbalanced counts) — no database writes
- `Import()` calls `PreviewImport()`, then:
  1. Creates missing accounts via `AccountService` (skips existing)
  2. Rejects unbalanced vouchers (never persists invalid bookkeeping)
  3. Converts SIE amount convention (positive = debit, negative = credit) to debit/credit line items
  4. Persists via `VoucherService.CreateVoucher()` + `LineItemService.CreateLineItem()`

The `SIEImportService` constructor wires `VoucherService`, `AccountService`, and `LineItemService` together because creating a voucher and its line items are separate writes that need coordinated orchestration.

## Plan Tiers & Feature Gating

### Plans

| Plan      | Price      | AI Agent | Invoicing | Notes                         |
|-----------|------------|----------|-----------|-------------------------------|
| `free`    | 0 kr       | Yes      | Yes       | 30-day trial, full access     |
| `mini`    | 99 kr/mån  | No       | No        | Core bookkeeping only         |
| `starter` | 199 kr/mån | Yes      | Yes       | Up to 100 transactions        |
| `growth`  | 399 kr/mån | Yes      | Yes       | Unlimited, priority support   |

### Middleware chain

```
AuthMiddleware → CompanyMiddleware → RequirePlan → Handler
```

1. **CompanyMiddleware** loads the company, checks trial expiration (30-day window for `free` plan → `TRIAL_EXPIRED`), checks payment status (`past_due` → `PAYMENT_PAST_DUE`), and sets `companyPlan` in the Gin context. Stripe routes are exempt from trial/payment checks so users can always reach the payment page.
2. **RequirePlan(allowedPlans ...string)** reads `companyPlan` from context. Returns `402 FEATURE_NOT_IN_PLAN` if the plan isn't in the allow-list.

**Gated routes:**
- `/agent/*` → `RequirePlan("free", "starter", "growth")` — Mini blocked
- `/invoices/*`, `/customers/*` → `RequirePlan("free", "starter", "growth")` — Mini blocked

The sidebar hides gated tabs in the UI so Mini users don't see features they can't access.

## Stripe Integration

**File:** `stripe_service.go`, `stripe_handler.go`

### Checkout flow

1. Frontend calls `POST /stripe/checkout` with the desired plan
2. Backend creates a Stripe Checkout Session with `company_id` in subscription metadata
3. User completes payment on Stripe's hosted page
4. Stripe fires `checkout.session.completed` webhook → backend processes it

### Webhook defense-in-depth

1. **Signature verification** — handler verifies Stripe webhook signature before passing to service
2. **Re-fetch from Stripe API** — even with a valid signature, the service re-fetches the subscription directly from Stripe. A forged payload with a fake subscription ID will 404.
3. **Company ID from metadata** — extracted from subscription metadata, not from the webhook payload body
4. **Price mapping with rejection** — `priceToPlan(priceID)` returns empty string for unknown prices; callers reject rather than defaulting to a paid tier

### Price → Plan mapping

| Environment | Mini | Starter | Growth |
|-------------|------|---------|--------|
| Test | `price_MINI_TEST_PLACEHOLDER` | `price_1TKFEmF27XxV0OojLLuaOAfg` | `price_1TKFFGF27XxV0Ooj3uaT41ho` |
| Live | `price_1TMRaZFQgVJ3jY3cOXaxsp29` | `price_1TKG9DFQgVJ3jY3cTQwpxKd7` | `price_1TKG9HFQgVJ3jY3cQWf0Tq16` |

### Status mapping

- `active` → plan active
- `past_due` → blocked by CompanyMiddleware (402)
- `canceled` / `unpaid` → reverts to `plan = "free"`, trial check kicks in
- `trialing` → Stripe-side trial (not currently used; we use our own 30-day trial)

## Organisationsnummer Validation

Swedish org numbers are validated on both sides with a Luhn checksum and normalized to `XXXXXX-XXXX` format.

**Frontend** (`lib/utils/orgNumber.ts`):
- `validateOrgNumber()` — returns Swedish error message or null (empty, wrong length, bad checksum)
- `formatOrgNumber()` — auto-formats input as the user types
- `luhnValid()` — standard Luhn algorithm

**Backend** (`company_service.go`, `normalizeOrgNumber()`):
- Strips non-digits, validates length = 10, validates Luhn checksum
- Returns canonical `XXXXXX-XXXX` form
- Called on both `CreateCompany()` and `UpdateCompany()`

Both layers validate independently — frontend for UX, backend for defense-in-depth.

## SKV 4700 Momsdeklaration PDF

**File:** `vat_pdf_handler.go`

**Endpoint:** `GET /reports/vat/pdf?from_date=YYYY-MM-DD&to_date=YYYY-MM-DD`

Generates a PDF matching the Skatteverket 4700 form (Swedish VAT declaration). VAT entries from `ReportService.GetVATReport()` are mapped to SKV boxes:

| Box | Description | Source |
|-----|-------------|--------|
| 05 | Försäljning som beskattas i Sverige (ex moms) | Sum of sales at 25% + 12% + 6% |
| 10 | Utgående moms 25% | VAT amount at 25% rate |
| 11 | Utgående moms 12% | VAT amount at 12% rate |
| 12 | Utgående moms 6% | VAT amount at 6% rate |
| 42 | Övrig försäljning m.m. (momsfri) | Sales at 0% rate |
| 48 | Ingående moms att dra av | Total input VAT (purchases) |
| 49 | Moms att betala / få tillbaka | Net VAT (positive = pay, negative = refund) |

The PDF includes sections A (taxable sales), B (output VAT), H (input VAT), and a final net amount with conditional labeling.

## Cloud SQL & Database Connectivity

### Development
- PostgreSQL 15 in Docker, exposed on port **5433** (not 5432)
- Connection: `postgres://postgres:postgres@localhost:5433/bookkeeping?sslmode=disable`

### Production
- **Google Cloud SQL** instance `esbio-app:europe-north1:esbio-db`
- Backend connects via Cloud SQL Unix socket (`/cloudsql/esbio-app:europe-north1:esbio-db`), configured through the `DATABASE_URL` env var on Cloud Run
- **Hardened** (2026-04-15):
  - `authorizedNetworks: []` — no internet access to DB
  - `sslMode: ENCRYPTED_ONLY` — unencrypted connections rejected
  - Public IP still present (Cloud SQL requires at least one of public/private/PSC) but unreachable from outside
- Full private IP migration (VPC + Serverless VPC Access connector) deferred due to cost (~$36/mo)

### Connection pooling
- Max open connections: 25
- Max idle connections: 5
- Ping validation on creation
- Deferred close on graceful shutdown

## Tink Bank Feed (Open Banking)

Esbio integrates with [Tink](https://tink.com) (Visa) for Swedish bank account connectivity via PSD2 Open Banking.

**Files:**
- `tink_client.go` — low-level Tink API wrapper (OAuth token exchange, accounts, transactions)
- `tink_service.go` — business logic (connect, sync, import, categorization, AES-256-GCM token encryption)
- `bank_handler.go` — HTTP endpoints
- Repositories: `bank_connection_repository.go`, `bank_account_repository.go`, `bank_transaction_repository.go`, `categorization_rule_repository.go`, `tink_oauth_state_repository.go`
- Frontend: `app/bank/page.tsx`, `app/bank/callback/page.tsx`, `lib/api/bank.ts`

### Connect flow

```
User clicks "Koppla bank"
    → Backend generates CSRF state token (stored in tink_oauth_state, 15min TTL)
    → Returns Tink Link URL
    → User redirects to link.tink.com, authenticates with bank
    → Tink redirects to /bank/callback?code=...&state=...
    → Frontend POSTs state+code to backend (auth only, no company middleware)
    → Backend validates state, exchanges code for access token
    → Encrypts token (AES-256-GCM), stores connection + accounts
    → Frontend redirects to /bank?connected=true
```

### Sync flow

Manual sync (user clicks "Synka"):
1. Decrypt stored access token
2. Fetch transactions from Tink API for each bank account
3. Bulk upsert into `bank_transactions` (deduplication via `tink_transaction_id` UNIQUE)
4. Auto-suggest accounts using `bank_categorization_rules` pattern matching
5. Update account balances and `last_synced_at`

### Transaction import

When user imports a bank transaction as a voucher:
1. Create voucher with 2 line items: bank BAS account (e.g. 1930) ↔ target account
2. Positive amount = income (debit bank, credit target); negative = expense (credit bank, debit target)
3. Mark transaction as `booked`, link to voucher
4. Upsert categorization rule from merchant/description → account mapping (learning)

### Categorization rules

The system learns from user imports. When a transaction is imported with account 6110 and merchant "Telia", a rule is created: `match_pattern: "Telia", match_type: "contains", account_no: 6110`. Future syncs auto-suggest 6110 for Telia transactions.

### Token security

- Access tokens encrypted at rest with AES-256-GCM (`TINK_ENCRYPTION_KEY` env var, 32 bytes hex)
- Dev mode: if no key configured, tokens stored in plaintext with warning log
- OAuth state tokens are single-use (DELETE on read) with 15-minute expiry

### Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/bank/connect` | auth+company | Returns Tink Link URL |
| POST | `/bank/callback` | auth only | Processes OAuth callback (companyID from state token) |
| GET | `/bank/connections` | auth+company | List connections |
| GET | `/bank/accounts` | auth+company | List bank accounts |
| GET | `/bank/transactions` | auth+company | List transactions (query: status, limit, offset) |
| POST | `/bank/sync/:connectionId` | auth+company | Manual sync |
| PUT | `/bank/accounts/:id/map` | auth+company | Map to BAS account |
| POST | `/bank/transactions/:id/import` | auth+company | Create voucher from transaction |
| POST | `/bank/transactions/:id/skip` | auth+company | Mark as skipped |
| DELETE | `/bank/connections/:id` | auth+company | Disconnect bank |

### Plan gating

Bank feed requires `RequirePlan("free", "starter", "growth")` — Mini plan blocked. Sidebar hides the Bank tab for Mini users.

## Company Selection Persistence

`CompanyContext` persists the selected company across page reloads and external redirects (e.g. Tink OAuth) using `localStorage`:

- **On select:** `localStorage.setItem("selectedCompanyId", id)` + backend `company_id` cookie
- **On mount:** `useState` initializer reads localStorage synchronously before first render, creating a placeholder `{ company_id }` object. This prevents `ProtectedRoute` from redirecting to `/companies` before the API loads the full company data.
- **On logout:** `localStorage.removeItem("selectedCompanyId")`
- The `company_id` cookie is set as non-httpOnly (it's just an integer, not sensitive) so the backend `CompanyMiddleware` can also read it.
- `GET /companies/selected` endpoint reads the cookie and returns the full company object (used for verification).
