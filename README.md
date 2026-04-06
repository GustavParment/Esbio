# Esbio - Bookkeeping Application

A modern, secure bookkeeping application using the Swedish BAS account system.

## Tech Stack

- **Frontend**: Next.js 16, React 19, TypeScript, Tailwind CSS
- **Backend**: Go with Gin framework
- **Database**: PostgreSQL 15
- **Authentication**: JWT with httpOnly cookies

## Quick Start

### Start Everything

```bash
./start.sh
```

This will:
1. Start the PostgreSQL Docker container (if not running)
2. Start the Go backend server on `:8080`
3. Start the Next.js frontend on `:3000`

### Stop Everything

```bash
./stop.sh
```

This will stop both the backend and frontend. You'll be prompted whether to stop the PostgreSQL container.

## Manual Setup

If you prefer to start services manually:

### 1. Start PostgreSQL

```bash
cd server
docker-compose up -d
```

### 2. Start Backend

```bash
cd server/cmd
go run api/main.go
```

### 3. Start Frontend

```bash
cd client
npm run dev
```

## Access URLs

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **Database**: postgres://localhost:5433

## Default Credentials

After registration, all users default to the "Bookkeeper" role. Admins can change user roles later.

## Project Structure

```
Esbio/
├── client/              # Next.js frontend
│   ├── app/            # Next.js App Router pages
│   ├── components/     # React components
│   ├── lib/            # API client & utilities
│   └── types/          # TypeScript definitions
├── server/             # Go backend
│   ├── cmd/api/        # Main application
│   │   └── internal/   # Handlers, middleware, services
│   ├── docker-compose.yml
│   └── init.sql        # Database migrations
├── start.sh            # Start all services
└── stop.sh             # Stop all services
```

## Features

- ✅ Secure httpOnly cookie authentication
- ✅ Mobile-responsive design
- ✅ Swedish BAS account system
- ✅ Double-entry bookkeeping
- ✅ Voucher management
- ✅ Account management
- ✅ User management with roles

## Security

- JWT tokens stored in httpOnly cookies (protected from XSS attacks)
- No localStorage usage for sensitive data
- CORS configured for localhost development
- Password hashing with bcrypt
