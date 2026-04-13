# Invoicing Module — Fas 1 Test Guide

Manual test steps for the customer registry and invoice settings features.
Run against local dev backend or prod after deploy.

---

## Prerequisites

1. Backend running (local: `go build -o /tmp/esbio-api ./api && cd /tmp && JWT_SECRET=... /tmp/esbio-api` or prod)
2. Frontend running (`cd client && npm run dev` or prod at `esbio.se`)
3. Docker postgres running locally (`cd server && docker compose up -d`)
4. Dev DB has the new tables (run the SQL from `init.sql` or migrations 010+011 if fresh DB)
5. Logged in with a user that has a selected company

---

## Test 1: Sidebar navigation

1. Log in at `/auth/login`
2. Select a company
3. Verify the sidebar shows two new items between "Verifikat" and "Konton":
   - **Fakturor** (invoice icon)
   - **Kunder** (person icon)
4. Click each — they should navigate to `/invoices` and `/customers` respectively
5. Active state (highlighted) should work when on those pages

**Pass criteria:** Both nav items visible, clickable, correct icons, active state works.

---

## Test 2: Customer list (empty state)

1. Navigate to `/customers`
2. With no customers created yet, you should see:
   - Header: "Kunder" + "Hantera ditt kundregister"
   - "Inga kunder registrerade" message
   - "Skapa din första kund" button
3. Search box should be visible but filtering an empty list

**Pass criteria:** Empty state renders cleanly, CTA button links to `/customers/new`.

---

## Test 3: Create a customer

1. Click "+ Ny kund" or navigate to `/customers/new`
2. Fill in:
   - Kundnamn: `Testföretag AB` (required — try submitting without it, should show error)
   - Org.nummer: `556123-4567`
   - E-post: `info@testforetag.se`
   - Telefon: `08-123 456`
   - Betalningsvillkor: `30` (default)
   - Adressrad 1: `Storgatan 1`
   - Postnummer: `111 22`
   - Ort: `Stockholm`
   - Land: `Sverige` (pre-filled)
3. Click "Skapa kund"
4. Should redirect to `/customers/<id>` with the customer detail view

**Pass criteria:** Customer created, all fields saved, redirect to detail page.

---

## Test 4: Customer detail view

1. On the customer detail page from Test 3, verify:
   - Name in header: "Testföretag AB"
   - All fields displayed correctly in the grid
   - "Redigera" and "Radera" buttons visible
2. Click "Redigera" — fields become editable inline
3. Change the phone number to `08-999 888`
4. Click "Spara"
5. Verify the updated phone number is displayed
6. Click "Avbryt" during edit — changes should be discarded

**Pass criteria:** Detail view shows all data, inline edit works, save persists, cancel discards.

---

## Test 5: Customer list (with data)

1. Navigate to `/customers`
2. The customer from Test 3 should appear in the table
3. Verify columns: Namn, Org.nr, E-post, Ort, Betalningsvillkor, Visa-link
4. Search for "testföretag" — should filter to show only matching customer
5. Search for "nonexistent" — should show "Inga kunder matchar din sökning"
6. Clear search — all customers visible again
7. "Visar X av Y kunder" counter at bottom should be accurate

**Pass criteria:** Table renders, search filters client-side, counter updates.

---

## Test 6: Delete customer

1. Create a throwaway customer (name: "Delete Me")
2. Go to the detail page
3. Click "Radera"
4. Confirm dialog should appear
5. Confirm — should redirect to `/customers` and customer is gone
6. Cancel — should stay on detail page, customer untouched

**Pass criteria:** Delete with confirmation works, redirect after delete.

---

## Test 7: Invoice settings (first visit)

1. Navigate to `/invoices/settings` (via sidebar "Fakturor" → "Inställningar" button)
2. On first visit for a company, default settings should auto-create:
   - F-skattetext: "Godkänd för F-skatt"
   - Betalningsvillkor: 30 dagar
   - Standardintäktskonto: 3010
   - Standardbetalningskonto: 1930
   - Nästa fakturanummer: 1
3. All payment fields (bankgiro, plusgiro, swish, IBAN, BIC) should be empty

**Pass criteria:** Settings page loads with defaults, no errors.

---

## Test 8: Update invoice settings

1. On the invoice settings page, fill in:
   - Bankgiro: `1234-5678`
   - Plusgiro: `12 34 56-7`
   - Swish: `1234567890`
   - F-skattetext: `Godkänd för F-skatt`
   - Sidfot: `Tack för er betalning!`
2. Click "Spara inställningar"
3. Green success message should appear: "Inställningar sparade"
4. Refresh the page — values should persist
5. Change betalningsvillkor to 15, save, verify it sticks

**Pass criteria:** All fields save and persist across page loads.

---

## Test 9: Invoices landing page

1. Navigate to `/invoices`
2. Should show placeholder state:
   - "Fakturering kommer snart" heading
   - Two CTA buttons: "Konfigurera betalningsuppgifter" → `/invoices/settings`, "Hantera kunder" → `/customers`
3. Both buttons should work

**Pass criteria:** Placeholder renders, links work. (This page will be replaced in Fas 2 with the actual invoice list.)

---

## Test 10: API endpoint verification (curl)

If you want to verify the backend independently of the frontend:

```bash
MAC=localhost:8080   # or 192.168.0.36:8080 for LAN access

# Login and get cookies
curl -sS -c /tmp/test.cookies -X POST http://$MAC/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"YOUR_EMAIL","password":"YOUR_PASSWORD"}'

# Select company
curl -sS -b /tmp/test.cookies -c /tmp/test.cookies -X POST http://$MAC/api/v1/companies/select \
  -H "Content-Type: application/json" -d '{"company_id":YOUR_ID}'

# Create customer
curl -sS -b /tmp/test.cookies -X POST http://$MAC/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"API Test AB","org_number":"556000-1234","email":"api@test.se","payment_terms_days":30}'
# → 201 with customer JSON

# List customers
curl -sS -b /tmp/test.cookies http://$MAC/api/v1/customers
# → array of customers

# Search customers
curl -sS -b /tmp/test.cookies "http://$MAC/api/v1/customers/search?q=API"
# → filtered array

# Get invoice settings (auto-creates defaults)
curl -sS -b /tmp/test.cookies http://$MAC/api/v1/invoices/settings
# → settings JSON with defaults

# Update invoice settings
curl -sS -b /tmp/test.cookies -X PUT http://$MAC/api/v1/invoices/settings \
  -H "Content-Type: application/json" \
  -d '{"bankgiro":"1234-5678","f_skatt_text":"Godkänd för F-skatt"}'
# → updated settings
```

---

## Test 11: Company isolation

1. Create two users with two different companies
2. As user A, create a customer
3. As user B, list customers — user A's customer should NOT appear
4. As user B, try `GET /customers/<user-A-customer-id>` — should 404 or the company check should prevent access

**Pass criteria:** Customer data is fully isolated per company.

---

## Edge cases to verify

- Creating a customer with only the name field (all others empty) — should work
- Creating a customer with a very long name (200 chars) — should work
- Payment terms of 0 — should default to 30
- Special characters in customer name (å, ä, ö, &, quotes) — should save and display correctly
- Invoice settings for a company that doesn't have a settings row yet — should auto-create defaults on first GET

---

## Known limitations (Fas 1)

- No invoice creation yet (Fas 2)
- No PDF generation (Fas 3)
- Invoice list page is a placeholder
- Customer deletion does not check for existing invoices (will add FK constraint in Fas 2)
- No pagination on customer list (fine for <100 customers, revisit if needed)
