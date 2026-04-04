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
3. Log in at [http://localhost:3000/auth/login](http://localhost:3000/auth/login)
4. You'll land on the dashboard

## Environment Variables

### Backend (`server/cmd/.env`)

| Variable            | Default                                                          | Description              |
|---------------------|------------------------------------------------------------------|--------------------------|
| `DATABASE_URL`      | `postgres://postgres:postgres@localhost:5433/bookkeeping?sslmode=disable` | PostgreSQL connection string |
| `JWT_SECRET`        | `your-secret-key-change-this-in-production`                      | JWT signing key          |
| `SERVER_PORT`       | `:8080`                                                          | Server port              |
| `ANTHROPIC_API_KEY` | (empty)                                                          | Required for AI agent    |

### Frontend (`client/.env.local`)

| Variable             | Default                                  | Description      |
|----------------------|------------------------------------------|------------------|
| `NEXT_PUBLIC_API_URL`| `http://localhost:8080/api/v1`           | Backend API URL  |

## Network Access

To access Eskio from other devices on your local network, see [network-access-setup.md](./network-access-setup.md).

## Project Structure

```
Eskio/
├── client/                  Next.js frontend
│   ├── app/                 Pages (App Router)
│   ├── components/          React components
│   ├── lib/api/             Typed API client
│   ├── lib/contexts/        Auth context
│   └── types/               TypeScript interfaces
├── server/                  Go backend
│   ├── cmd/api/main.go      Entry point
│   ├── cmd/api/internal/
│   │   ├── handlers/        HTTP handlers (incl. agent)
│   │   ├── service/         Business logic (incl. agent, scheduler)
│   │   ├── repository/      Database access (incl. scheduled tasks, messages)
│   │   ├── domain/          Data models
│   │   ├── dto/             Request/response DTOs
│   │   ├── middleware/      Auth & CORS
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
