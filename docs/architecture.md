# Architecture Overview

Esbio is a full-stack Swedish bookkeeping application built around double-entry accounting principles using the BAS (Bas Kontoplan) chart of accounts.

## Tech Stack

| Layer        | Technology                          |
|--------------|-------------------------------------|
| Frontend     | Next.js 16, React 19, TypeScript    |
| Styling      | Tailwind CSS 4                      |
| Backend      | Go with Gin web framework           |
| Database     | PostgreSQL 15 (Docker)              |
| Auth         | JWT (HS256) in httpOnly cookies     |
| PDF          | go-pdf/fpdf                         |

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
5. Instantiate services with their repositories
6. Instantiate handlers with their services
7. Register routes with Gin router
8. Start server

## Frontend Structure

The client uses the **Next.js App Router** with client-side rendering for protected routes.

```
app/
├── layout.tsx              Root layout with AuthProvider
├── page.tsx                Redirects to /auth/login
├── auth/                   Login & register pages
├── companies/              Company selector ("Mina Företag") + create company
├── dashboard/              Main dashboard
├── vouchers/               Voucher CRUD + corrections
├── accounts/               Account CRUD + ledger
├── reports/                Income statement, balance sheet, VAT report
├── agent/                  AI assistant chat + scheduled tasks
├── settings/               Company settings (edits company info)
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
- **12 tools** — wrapping existing services (voucher CRUD incl. corrections and search, reports, account ledger, scheduled tasks)
- **SchedulerService** — background goroutine checking for due tasks every 60 seconds
- **ScheduledTaskService** — CRUD for recurring monthly tasks with next-run calculation
- **System prompt** — Swedish-language, BAS-aware, with security rules against prompt injection
- **Error sanitization** — database errors are never exposed to users; sanitized messages are returned instead

**Security:** Company ID is enforced server-side on all tool calls. All data is scoped to the selected company. See [ester-ai.md](./ester-ai.md) for full security documentation.

## Security

- Passwords hashed with **bcrypt**
- JWT stored in **httpOnly cookies** (not accessible via JavaScript)
- CORS whitelist for allowed origins
- **CompanyMiddleware** verifies company ownership on every request
- All data queries filter by `company_id` for strict data isolation between companies
- Role-based middleware (`RequireRole("Admin")`) for sensitive endpoints
- Input validation via `go-playground/validator`
- Foreign key constraints with `ON DELETE RESTRICT` / `CASCADE`
