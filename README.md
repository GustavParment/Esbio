# Esbio - Svenskt Bokföringssystem

Modernt, säkert bokföringsprogram enligt svenska BAS-kontoplanen.

**Live:** [https://esbio.se](https://esbio.se)

## Tech Stack

- **Frontend**: Next.js 16, React 19, TypeScript, Tailwind CSS
- **Backend**: Go with Gin framework
- **Database**: PostgreSQL 15 (Cloud SQL)
- **Hosting**: Google Cloud Run (europe-north1)
- **DNS**: Cloudflare
- **Authentication**: JWT with httpOnly cookies (SameSite=None, Secure)

## Branches

- **`main`** — Produktion. API pekar mot `https://api.esbio.se`
- **`dev`** — Lokal utveckling. API pekar mot `http://localhost:8080`

## Local Development

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

### Manual Setup

```bash
# 1. Database
cd server && docker-compose up -d

# 2. Backend
cd server/cmd && go run api/main.go

# 3. Frontend
cd client && npm run dev
```

### Local URLs

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **Database**: postgres://localhost:5433

## Production

- **Frontend**: https://esbio.se (Cloud Run: esbio-frontend)
- **Backend API**: https://api.esbio.se (Cloud Run: esbio-backend)
- **Database**: Cloud SQL (esbio-db, europe-north1)

### Deploy

```bash
# Frontend
cd client && gcloud run deploy esbio-frontend --source . --region europe-north1 --port 3000 --allow-unauthenticated

# Backend
cd server && gcloud run deploy esbio-backend --source . --region europe-north1 --allow-unauthenticated
```

### Environment Variables (Cloud Run Backend)

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | JWT signing secret (required) |
| `DATABASE_URL` | Cloud SQL connection string |
| `SERVER_PORT` | Server port (default: `:8080`) |
| `GEMINI_API_KEY` | Google Gemini API key (for Ester AI & receipt scanning) |
| `STRIPE_SECRET_KEY` | Stripe secret key for subscription billing |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret |
| `FRONTEND_URL` | Frontend URL for Stripe redirects (default: `http://localhost:3000`) |
| `COOKIE_SAMESITE` | Cookie SameSite policy (`none` or `lax`) |
| `COOKIE_SECURE` | Cookie Secure flag (`true` or `false`) |

## Project Structure

```
Esbio/
├── client/                  # Next.js frontend
│   ├── app/                # App Router pages
│   ├── components/         # React components
│   ├── lib/                # API client, contexts & utilities
│   └── types/              # TypeScript definitions
├── server/                  # Go backend
│   ├── cmd/api/            # Main application
│   │   └── internal/
│   │       ├── handlers/   # HTTP handlers
│   │       ├── middleware/  # Auth, CORS, rate limiting, security headers
│   │       ├── service/    # Business logic
│   │       ├── repository/ # Database queries
│   │       ├── config/     # Configuration
│   │       └── routes/     # Route definitions
│   ├── docker-compose.yml
│   └── init.sql            # Database migrations
├── start.sh
└── stop.sh
```

## Features

- Secure httpOnly cookie authentication
- Mobile-responsive design
- Swedish BAS account system
- Double-entry bookkeeping
- Voucher management with PDF export
- AI receipt scanning (Gemini Vision — photo to voucher)
- Account management & ledger
- Financial reports (income statement, balance sheet, VAT with input/output moms & net calculation)
- SIE file export
- Ester AI assistant (Gemini-powered)
- Stripe subscription billing (Starter 199kr/mån, Tillväxt 399kr/mån)
- Self-service account deletion
- Multi-company support
- Dark mode

## Stripe Billing

Subscription billing via Stripe Checkout (hosted payment page — no card data on our servers).

**Plans:**
| Plan | Price | Features |
|------|-------|----------|
| Free (trial) | 0 kr / 30 dagar | Full access during trial |
| Starter | 199 kr/mån | 100 transactions/mån, AI, receipt scanning, VAT reports |
| Tillväxt | 399 kr/mån | Unlimited transactions, annual reports, priority support |

**How it works:**
1. User clicks upgrade → Stripe Checkout session created with `company_id` in metadata
2. User pays on Stripe's hosted page → redirected to `/upgrade/success`
3. Stripe sends `checkout.session.completed` webhook → backend fetches subscription → updates `plan`, `plan_status` in DB
4. Subscription lifecycle events (`updated`, `deleted`) also handled via webhook
5. `past_due` subscriptions blocked with HTTP 402 in middleware

**Endpoints:**
- `POST /stripe/checkout` — Create Checkout session (auth required)
- `POST /stripe/portal` — Open Stripe Customer Portal for plan management/cancellation (auth required)
- `POST /stripe/webhook` — Stripe webhook receiver (no auth)

**Local development:**
```bash
stripe login
stripe listen --forward-to localhost:8080/api/v1/stripe/webhook
```
The webhook listener starts automatically with `./start.sh` if the Stripe CLI is installed.

**Production setup:**
1. Create products/prices in Stripe Dashboard
2. Set `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET` in Cloud Run env vars
3. Register webhook endpoint: `https://api.esbio.se/api/v1/stripe/webhook`
4. Configure Customer Portal in Stripe Dashboard (allow cancel + plan switching)

## Security

- JWT tokens in httpOnly cookies (XSS protected)
- Rate limiting: 100 req/min global, 10 req/min on auth endpoints
- Security headers: HSTS, X-Frame-Options, X-Content-Type-Options
- Error sanitization: internal errors logged server-side only
- bcrypt password hashing
- CORS whitelist
- Parameterized SQL queries (injection protected)
- Company data isolation via middleware
- Role-based access control (RBAC) on all write operations

## Roles & Permissions

Three roles: **Admin**, **Bookkeeper**, **Manager**. New users default to Bookkeeper.

| Operation | Admin | Bookkeeper | Manager |
|-----------|:-----:|:----------:|:-------:|
| View data (accounts, vouchers, reports) | Yes | Yes | Yes |
| Create/edit accounts | Yes | Yes | No |
| Create/edit vouchers & line items | Yes | Yes | No |
| Delete accounts & line items | Yes | No | No |
| Update/delete vouchers | Yes | No | No |
| Create correction vouchers | Yes | No | No |
| Create/delete users | Yes | No | No |
| Change user roles | Yes | No | No |
| Read/update own profile | Yes | Yes | Yes |
| Company CRUD (own companies) | Yes | Yes | Yes |
| Export reports (SIE, PDF) | Yes | Yes | Yes |
