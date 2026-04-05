# API Reference

All endpoints are prefixed with `/api/v1`. Authenticated endpoints require a valid JWT cookie. Most endpoints also require a `company_id` cookie (set via company selection) and the CompanyMiddleware verifies that the authenticated user owns the selected company.

## Authentication

| Method | Endpoint              | Description            | Auth |
|--------|-----------------------|------------------------|------|
| POST   | `/auth/register`      | Register a new user    | No   |
| POST   | `/auth/login`         | Login, sets JWT cookie | No   |
| POST   | `/auth/logout`        | Clears JWT cookie      | No   |
| POST   | `/auth/refresh`       | Refresh JWT token      | No   |
| GET    | `/auth/me`            | Get current user       | Yes  |

### POST /auth/register

```json
{
  "name": "Anna Svensson",
  "email": "anna@example.com",
  "password": "minst8tecken",
  "role": "Bookkeeper"
}
```

Roles: `Admin`, `Bookkeeper`, `Manager`. Defaults to `Bookkeeper`.

### POST /auth/login

```json
{
  "email": "anna@example.com",
  "password": "minst8tecken"
}
```

Response sets an httpOnly cookie named `token` (7-day expiry). After login, the user should be directed to `/companies` to select a company before accessing the app.

---

## Companies

| Method | Endpoint                  | Description                     | Auth |
|--------|---------------------------|---------------------------------|------|
| GET    | `/companies`              | List current user's companies   | Yes  |
| POST   | `/companies`              | Create a new company            | Yes  |
| POST   | `/companies/select`       | Set company_id cookie           | Yes  |
| PUT    | `/companies/:id`          | Update company                  | Yes  |
| DELETE | `/companies/:id`          | Delete company                  | Yes  |

### POST /companies

```json
{
  "company_name": "Acme AB",
  "org_number": "556123-4567",
  "plan": "free"
}
```

### POST /companies/select

Sets a `company_id` httpOnly cookie that scopes all subsequent requests to the selected company.

```json
{
  "company_id": 1
}
```

### PUT /companies/:id

```json
{
  "company_name": "Acme AB (updated)",
  "org_number": "556123-4567",
  "plan": "pro"
}
```

---

## Users

| Method | Endpoint               | Description        | Auth  |
|--------|------------------------|--------------------|-------|
| POST   | `/users`               | Create user        | Yes   |
| GET    | `/users/:id`           | Get user by ID     | Yes   |
| GET    | `/users/email/:email`  | Get user by email  | Yes   |
| PUT    | `/users/:id`           | Update user        | Yes   |
| DELETE | `/users/:id`           | Delete user        | Yes   |

---

## Accounts

| Method | Endpoint                       | Description              | Auth |
|--------|--------------------------------|--------------------------|------|
| POST   | `/accounts`                    | Create account           | Yes  |
| GET    | `/accounts`                    | Get all accounts         | Yes  |
| GET    | `/accounts/:accountNo`         | Get account by number    | Yes  |
| GET    | `/accounts/:accountNo/ledger`  | Get account ledger       | Yes  |
| GET    | `/accounts/group/:group`       | Get accounts by group    | Yes  |
| PUT    | `/accounts/:accountNo`         | Update account           | Yes  |
| DELETE | `/accounts/:accountNo`         | Delete account           | Yes  |

### GET /accounts/:accountNo/ledger

Query parameters:
- `period` (optional) — filter by period, e.g. `2025-01`

Returns ledger entries with running balance. Corrected vouchers are excluded.

### Account Groups (BAS)

| Group | Category                    |
|-------|-----------------------------|
| 1     | Tillgangar (Assets)         |
| 2     | Eget kapital & skulder      |
| 3     | Intakter (Income)           |
| 4     | Material & varor            |
| 5     | Ovriga externa kostnader    |
| 6     | Personal                    |
| 7     | Avskrivningar               |
| 8     | Finansiella poster          |

---

## Vouchers

| Method | Endpoint                              | Description                    | Auth  |
|--------|---------------------------------------|--------------------------------|-------|
| POST   | `/vouchers`                           | Create voucher                 | Yes   |
| GET    | `/vouchers`                           | Get all vouchers               | Yes   |
| GET    | `/vouchers/periods`                   | Get all unique periods         | Yes   |
| GET    | `/vouchers/:id`                       | Get voucher with line items    | Yes   |
| GET    | `/vouchers/period/:period`            | Get vouchers by period         | Yes   |
| GET    | `/vouchers/company`                   | Get vouchers for selected company | Yes |
| GET    | `/vouchers/:id/validate`              | Check if debit = credit        | Yes   |
| POST   | `/vouchers/:id/correct`               | Create reversal voucher        | Yes   |
| POST   | `/vouchers/:id/correct-with-changes`  | Create correction with updates | Yes   |
| GET    | `/vouchers/:id/pdf`                   | Download voucher as PDF        | Yes   |
| PUT    | `/vouchers/:id`                       | Update voucher                 | Admin |
| DELETE | `/vouchers/:id`                       | Delete voucher                 | Admin |

### POST /vouchers

```json
{
  "date": "2025-01-15",
  "description": "Faktura #1234",
  "reference": "F-1234",
  "total_amount": 10000.00,
  "period": "2025-01",
  "lines": [
    { "account_no": 1930, "debit_amount": 10000, "credit_amount": 0, "tax_code": 25 },
    { "account_no": 3010, "debit_amount": 0, "credit_amount": 10000, "tax_code": 25 }
  ]
}
```

### GET /vouchers/:id/validate

```json
{
  "balanced": true,
  "total_debit": 10000.00,
  "total_credit": 10000.00,
  "difference": 0.00
}
```

### Voucher Corrections

**Simple reversal** (`POST /vouchers/:id/correct`): Creates a new voucher with debit/credit swapped. Marks the original as corrected.

**Correction with changes** (`POST /vouchers/:id/correct-with-changes`): Creates a new voucher with the provided updated values. Marks the original as corrected.

A voucher that has already been corrected cannot be corrected again.

---

## Line Items

| Method | Endpoint                           | Description                 | Auth |
|--------|------------------------------------|-----------------------------|------|
| POST   | `/lineitems`                       | Create line item            | Yes  |
| GET    | `/lineitems/:id`                   | Get line item by ID         | Yes  |
| GET    | `/lineitems/voucher/:voucherId`    | Get items for voucher       | Yes  |
| GET    | `/lineitems/account/:accountNo`    | Get items for account       | Yes  |
| PUT    | `/lineitems/:id`                   | Update line item            | Yes  |
| DELETE | `/lineitems/:id`                   | Delete line item            | Yes  |

### Business Rules

- Each line item must have either `debit_amount > 0` OR `credit_amount > 0`, never both.
- Both `voucher_id` and `account_no` must reference existing records.

---

## Reports

| Method | Endpoint                     | Description          | Auth |
|--------|------------------------------|----------------------|------|
| GET    | `/reports/income-statement`  | Income statement     | Yes  |
| GET    | `/reports/balance-sheet`     | Balance sheet        | Yes  |
| GET    | `/reports/vat`               | VAT report           | Yes  |

### GET /reports/income-statement

Query parameters:
- `from_date` — start date (YYYY-MM-DD)
- `to_date` — end date (YYYY-MM-DD)

Returns:
- Income entries (accounts 3000-3999)
- Expense entries (accounts 4000-8999)
- Total income, total expenses, net result
- Corrected vouchers are excluded

### GET /reports/balance-sheet

Query parameters:
- `as_of_date` — the date to calculate balances up to (YYYY-MM-DD)

Returns:
- Assets (group 1 accounts) with cumulative balances
- Equity & liabilities (group 2 accounts) with cumulative balances
- P&L net result (carried into equity side)
- Total assets, total equity & liabilities
- Corrected vouchers are excluded

```json
{
  "as_of_date": "2025-12-31",
  "assets": [
    { "account_no": 1930, "account_name": "Bankkonto", "balance": 50000.00 }
  ],
  "equity_liabilities": [
    { "account_no": 2010, "account_name": "Eget kapital", "balance": -30000.00 }
  ],
  "total_assets": 50000.00,
  "total_equity_liabilities": -50000.00,
  "net_result": -20000.00
}
```

### GET /reports/vat

Query parameters:
- `from_date` — start date (YYYY-MM-DD)
- `to_date` — end date (YYYY-MM-DD)

Returns VAT (moms) breakdown by tax rate for revenue accounts (3000-3999):
- Per tax code: sales amount, calculated VAT, tax rate label
- Total sales and total VAT across all rates
- Corrected vouchers are excluded

```json
{
  "period": { "from_date": "2025-01-01", "to_date": "2025-12-31" },
  "entries": [
    { "tax_code": 25, "tax_rate": "25%", "total_sales": 100000.00, "total_vat": 25000.00 },
    { "tax_code": 12, "tax_rate": "12%", "total_sales": 20000.00, "total_vat": 2400.00 }
  ],
  "total_sales": 120000.00,
  "total_vat": 27400.00
}
```

---

## Ester AI (Agent)

Ester AI is Eskio's intelligent bookkeeping assistant. She can create vouchers, look up accounts, generate reports, and schedule recurring tasks. See [ester-ai.md](./ester-ai.md) for full documentation including security model.

Requires the `GEMINI_API_KEY` environment variable to be set in `server/cmd/.env`. Ester is company-scoped: all tool calls use the selected company_id from the cookie.

| Method | Endpoint                      | Description              | Auth |
|--------|-------------------------------|--------------------------|------|
| POST   | `/agent/chat`                 | Send message to agent    | Yes  |
| GET    | `/agent/messages/:conversationId` | Get conversation history | Yes  |
| GET    | `/agent/tasks`                | List scheduled tasks     | Yes  |
| PUT    | `/agent/tasks/:id/toggle`     | Pause/resume a task      | Yes  |
| DELETE | `/agent/tasks/:id`            | Delete a scheduled task  | Yes  |

### POST /agent/chat

```json
{
  "message": "Skapa ett konsultarvode på 64 350 kr inkl moms",
  "conversation_id": ""
}
```

Response:
```json
{
  "response": "Jag har skapat verifikat #34 med konsultarvode...",
  "conversation_id": "conv-1-42"
}
```

The agent has access to the following tools:
- `get_voucher` — fetch a voucher by ID
- `get_vouchers_by_period` — list vouchers for a period (YYYY-MM)
- `get_company_vouchers` — list vouchers for the selected company
- `create_voucher` — create a voucher with line items
- `correct_voucher` — create a reversal + new correct voucher
- `get_account_ledger` — account transaction history with running balance
- `search_accounts` — search the BAS chart of accounts
- `get_income_statement` — income statement for a date range
- `get_balance_sheet` — balance sheet as of a date
- `create_scheduled_task` — schedule a recurring monthly voucher
- `list_scheduled_tasks` — list a user's scheduled tasks

### Scheduled Tasks

Scheduled tasks are recurring jobs that the agent executes automatically. A background scheduler checks for due tasks every 60 seconds.

```json
{
  "task_id": 1,
  "company_id": 1,
  "description": "Konsultarvode månatlig",
  "prompt": "Skapa konsultarvode på 64350 kr inkl moms",
  "template_voucher_id": 26,
  "day_of_month": 30,
  "active": true,
  "last_run_at": "2025-06-30T08:00:00",
  "next_run_at": "2025-07-30T08:00:00"
}
