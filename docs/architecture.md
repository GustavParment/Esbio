# Architecture Overview

Eskio is a full-stack Swedish bookkeeping application built around double-entry accounting principles using the BAS (Bas Kontoplan) chart of accounts.

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
│  Middleware   │  Auth (JWT), CORS, Role-based access
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
4. Instantiate repositories (User, Account, LineItem, Voucher, Report)
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
├── dashboard/              Main dashboard
├── vouchers/               Voucher CRUD + corrections
├── accounts/               Account CRUD + ledger
├── reports/                Income statement, balance sheet, VAT report
├── agent/                  AI assistant chat + scheduled tasks
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

- **Global state:** React Context for authentication (`AuthContext`)
- **Local state:** `useState` / `useEffect` in page components
- **No external state library** (no Redux, Zustand, etc.)

### API Client

A base `ApiClient` class (`lib/api/client.ts`) handles:
- JSON serialization
- `credentials: 'include'` for cookie-based auth
- Error handling with typed `ApiError`
- Generic response typing

Domain-specific modules (`auth.ts`, `vouchers.ts`, `accounts.ts`, etc.) expose typed methods that call the base client.

## Authentication Flow

```
1. User submits credentials at /auth/login
2. Server validates, creates JWT (7-day expiry)
3. Server sets token in httpOnly cookie
4. All subsequent API calls include the cookie automatically
5. AuthMiddleware validates the token on protected endpoints
6. On app load, AuthContext calls /auth/me to restore session
7. ProtectedRoute redirects to /auth/login if no session
```

## Ester AI (Bookkeeping Assistant)

Eskio includes **Ester AI**, an intelligent bookkeeping assistant powered by Gemini 2.5 Flash (Google).

```
User message → Agent Handler → Gemini API (with tool definitions)
                                      │
                              Gemini decides to call tools
                                      │
                              Agent executes tools against existing services
                              (user ID enforced server-side)
                                      │
                              Result returned to Gemini → response to user
```

**Key components:**
- **AgentService** — orchestrates Gemini API calls with a tool-use loop (max 10 iterations)
- **10 tools** — wrapping existing services (voucher CRUD, reports, account ledger, scheduled tasks)
- **SchedulerService** — background goroutine checking for due tasks every 60 seconds
- **ScheduledTaskService** — CRUD for recurring monthly tasks with next-run calculation
- **System prompt** — Swedish-language, BAS-aware, with security rules against prompt injection

**Security:** User ID is enforced server-side on all tool calls. See [ester-ai.md](./ester-ai.md) for full security documentation.

## Security

- Passwords hashed with **bcrypt**
- JWT stored in **httpOnly cookies** (not accessible via JavaScript)
- CORS whitelist for allowed origins
- Role-based middleware (`RequireRole("Admin")`) for sensitive endpoints
- Input validation via `go-playground/validator`
- Foreign key constraints with `ON DELETE RESTRICT` / `CASCADE`
