# Getting Started

## Prerequisites

- **Docker** — for running PostgreSQL
- **Go** 1.21+ — for the backend server
- **Node.js** 18+ and npm — for the frontend
- **Git**

## Quick Start

From the project root:

```bash
./start.sh
```

This starts all three services:
1. PostgreSQL container (Docker) on port **5433**
2. Go backend server on port **8080**
3. Next.js frontend on port **3000**

Open [http://localhost:3000](http://localhost:3000) in your browser.

To stop everything:

```bash
./stop.sh
```

## Manual Setup

### 1. Start the database

```bash
cd server
docker compose up -d
```

This creates a PostgreSQL 15 container and runs `init.sql` to set up the schema and seed the BAS chart of accounts.

**Connection details:**
- Host: `localhost`
- Port: `5433` (mapped from container's 5432)
- Database: `bookkeeping`
- User: `postgres`
- Password: `postgres`

### 2. Start the backend

```bash
cd server/cmd
go run api/main.go
```

The API server starts on `http://localhost:8080`. It reads configuration from `server/cmd/.env` (or environment variables).

### 3. Start the frontend

```bash
cd client
npm install   # first time only
npm run dev
```

The frontend starts on `http://localhost:3000`.

## First Steps

1. Go to [http://localhost:3000/auth/register](http://localhost:3000/auth/register)
2. Create an account (default role: Bookkeeper)
3. Check your email for a **verification link** (requires `RESEND_API_KEY`)
4. Click the link to verify your email — a welcome email is sent
5. Log in at [http://localhost:3000/auth/login](http://localhost:3000/auth/login) (blocked until email is verified)
6. You'll be redirected to the **company selector** at `/companies`
7. Create a new company (name, org number) or select an existing one
8. After selecting a company, you'll land on the dashboard
9. All data (vouchers, reports, etc.) is scoped to the selected company
10. To switch companies, use the company switch link in the sidebar

**Note:** If running without `RESEND_API_KEY`, you can manually verify users:
```sql
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -d bookkeeping -c "UPDATE users SET email_verified = true WHERE email = 'your@email.com';"
```

## Environment Variables

### Backend (`server/cmd/.env`)

| Variable            | Default                                                          | Description              |
|---------------------|------------------------------------------------------------------|--------------------------|
| `DATABASE_URL`      | `postgres://postgres:postgres@localhost:5433/bookkeeping?sslmode=disable` | PostgreSQL connection string |
| `JWT_SECRET`        | `your-secret-key-change-this-in-production`                      | JWT signing key          |
| `SERVER_PORT`       | `:8080`                                                          | Server port              |
| `GEMINI_API_KEY`    | (empty)                                                          | Required for Ester AI agent |
| `RESEND_API_KEY`    | (empty)                                                          | Required for email (verification, invoices, support) |
| `FRONTEND_URL`      | `http://localhost:3000`                                          | Used in email links (verify, reset) |
| `STRIPE_SECRET_KEY` | (empty)                                                          | Stripe subscription billing |
| `STRIPE_WEBHOOK_SECRET` | (empty)                                                      | Stripe webhook verification |

### Frontend (`client/.env.local`)

| Variable             | Default                                  | Description      |
|----------------------|------------------------------------------|------------------|
| `NEXT_PUBLIC_API_URL`| `http://localhost:8080/api/v1`           | Backend API URL  |

## Network Access

To access Esbio from other devices on your local network, see [network-access-setup.md](./network-access-setup.md).

## Project Structure

```
Esbio/
├── client/                  Next.js frontend
│   ├── app/                 Pages (App Router)
│   ├── components/          React components
│   ├── lib/api/             Typed API client
│   ├── lib/contexts/        Auth context
│   └── types/               TypeScript interfaces
├── server/                  Go backend
│   ├── cmd/api/main.go      Entry point
│   ├── cmd/api/internal/
│   │   ├── handlers/        HTTP handlers (incl. agent, invoice email, support)
│   │   ├── service/         Business logic (incl. agent, scheduler, email)
│   │   ├── repository/      Database access (incl. scheduled tasks, messages)
│   │   ├── domain/          Data models
│   │   ├── dto/             Request/response DTOs
│   │   ├── middleware/      Auth, CORS & Company middleware
│   │   ├── routes/          Route definitions
│   │   ├── auth/            JWT management
│   │   ├── config/          Configuration
│   │   └── database/        DB connection
│   ├── init.sql             Schema & seed data
│   └── docker-compose.yml   PostgreSQL container
├── docs/                    Documentation
├── start.sh                 Start all services
└── stop.sh                  Stop all services
```
