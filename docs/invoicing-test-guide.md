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

- Customer deletion does not check for existing invoices (FK RESTRICT on invoices.customer_id prevents orphans at DB level)
- No pagination on customer list (fine for <100 customers, revisit if needed)

---

# Fas 2 — Invoice CRUD + Auto-bokforing

## Test 12: Create a draft invoice

1. Ensure at least one customer exists (from Fas 1 tests)
2. Ensure invoice settings are configured (bankgiro, F-skatt from Test 8)
3. Navigate to `/invoices/new`
4. Select a customer from the dropdown
5. Set fakturadatum to today, betalningsvillkor 30
6. Add invoice lines:
   - Rad 1: "Konsulttjänst april", antal 10, enhet "tim", á-pris 950, moms 25%
   - Rad 2: "Resekostnader", antal 1, enhet "st", á-pris 1500, moms 25%
7. Verify live calculation:
   - Netto: (10×950 + 1×1500) = 11 000,00 kr
   - Moms 25%: 2 750,00 kr
   - Att betala: 13 750 kr (already whole krona, no rounding)
8. Click "Skapa faktura (utkast)"
9. Should redirect to invoice detail page showing status "Utkast"
10. Invoice number should be assigned (e.g. #1 for first invoice)
11. OCR number should be visible (Luhn check digit appended)

**Pass criteria:** Draft invoice created with correct totals, OCR number, and line items.

## Test 13: Invoice list with status tabs

1. Navigate to `/invoices`
2. The draft invoice from Test 12 should appear
3. Click "Utkast" tab — should show the invoice
4. Click "Skickade" tab — should show empty
5. Click "Alla" tab — back to full list
6. Verify columns: Nr, Kund (name, not ID), Datum, Förfaller, Status badge, Belopp

**Pass criteria:** List renders, tabs filter correctly, customer name resolved.

## Test 14: Finalize invoice (draft → sent, auto-bokforing)

1. Go to the draft invoice detail page
2. Click "Skicka faktura"
3. Confirm the dialog
4. Status should change to "Skickad"
5. A link to the sales voucher should appear ("Visa verifikat")
6. Click the voucher link — verify the voucher has:
   - Debit 1510 (Kundfordringar): 13 750,00 kr
   - Credit 3010 (Försäljning): 11 000,00 kr
   - Credit 2610 (Utg. moms 25%): 2 750,00 kr
   - Total balanced
7. Go back to invoice — action buttons should now show "Markera som betald" and "Makulera"

**Pass criteria:** Status transition, voucher auto-generated with correct BAS accounts and amounts.

## Test 15: Mark invoice as paid (sent → paid, payment voucher)

1. On the sent invoice detail page
2. Set betalningsdatum to today
3. Click "Markera som betald"
4. Status should change to "Betald"
5. A link to the payment voucher should appear
6. Click the voucher link — verify:
   - Debit 1930 (Bank): 13 750,00 kr
   - Credit 1510 (Kundfordringar): 13 750,00 kr
   - Balanced

**Pass criteria:** Payment recorded, voucher correct, 1510 zeroed out.

## Test 16: Cancel invoice (sent → cancelled, correction voucher)

1. Create and finalize a new invoice (repeat Test 12 + 14 with different data)
2. On the sent invoice detail page, click "Makulera faktura"
3. Confirm the dialog
4. Status should change to "Makulerad"
5. The original sales voucher should now be marked as corrected
6. Verify in /vouchers that a correction voucher was created (reversed amounts)

**Pass criteria:** Cancellation creates correction voucher, original voucher marked corrected.

## Test 17: Delete draft invoice

1. Create a new draft invoice
2. On the detail page, click "Radera utkast"
3. Confirm — should redirect to `/invoices`
4. The invoice should be gone from the list
5. Try to delete a sent invoice via API — should fail:
   ```bash
   curl -sS -b /tmp/test.cookies -X DELETE http://localhost:8080/api/v1/invoices/<sent_id>
   # → {"error":"only draft invoices can be deleted"}
   ```

**Pass criteria:** Draft deletable, non-draft protected.

## Test 18: Oresavrundning (Swedish rounding)

1. Create an invoice with lines that produce a non-whole total:
   - "Tjänst", antal 1, á-pris 99.50, moms 25%
   - Netto: 99.50, moms: 24.875 → total raw: 124.375
   - Rounded: 124 kr, avrundning: -0.375 → shown as "Öresavrundning: -0,38 kr"
2. Finalize the invoice
3. Check the voucher — it should include an extra line for account 3740 (Öresavrundning)
4. Verify the voucher is balanced despite the rounding

**Pass criteria:** Rounding calculated and booked correctly.

## Test 19: OCR number validation

1. Create invoice #1 → OCR should be "18" (1 + Luhn digit 8)
2. Create invoice #2 → OCR should be "26"
3. Create invoice #10 → OCR should be "109"
4. Verify manually: take the digits before the last one, apply Luhn, check digit matches

**Pass criteria:** OCR numbers follow Luhn algorithm.

## Test 20: Multiple VAT rates on one invoice

1. Create an invoice with mixed VAT rates:
   - "Konsulttjänst", 10000 kr, 25% moms
   - "Livsmedel", 5000 kr, 12% moms
   - "Tidning", 2000 kr, 6% moms
2. Finalize the invoice
3. Check the voucher has separate credit lines:
   - 2610 (25%): 2 500,00 kr
   - 2611 (12%): 600,00 kr
   - 2612 (6%): 120,00 kr
   - 3010 (Försäljning): 17 000,00 kr
   - 1510 (Kundfordringar): total including all VAT

**Pass criteria:** Each VAT rate gets its own account line in the voucher.

## Test 21: Invoice settings snapshot

1. Configure bankgiro to "1111-2222" in invoice settings
2. Create and finalize an invoice — it should show bankgiro "1111-2222"
3. Change bankgiro to "3333-4444" in settings
4. View the old invoice — it should still show "1111-2222" (snapshotted at creation)
5. Create a new invoice — it should show "3333-4444"

**Pass criteria:** Payment details are frozen at invoice creation time.

## Test 22: Company isolation for invoices

1. As user A (company A), create a customer and an invoice
2. As user B (company B), list invoices — A's invoice should NOT appear
3. As user B, try `GET /invoices/<A's invoice id>` — should get 403

**Pass criteria:** Full company-scoped isolation.

---

# Fas 3 — Invoice PDF

## Test 23: Download invoice PDF

1. Create and finalize an invoice with 2-3 line items
2. On the invoice detail page, click the PDF download button
3. The browser should download a `.pdf` file
4. Open the PDF and verify it contains:
   - Company name and org number at the top
   - "FAKTURA" heading
   - Invoice number, date, due date
   - OCR number
   - Customer name and address
   - Line items table with description, quantity, unit, unit price, VAT rate, amount
   - VAT summary (per rate)
   - Totals: netto, moms, öresavrundning (if any), att betala
   - Payment details: bankgiro/plusgiro, OCR
   - F-skattetext at the bottom

**Pass criteria:** Professional Swedish invoice PDF with all required information.

## Test 24: PDF with oresavrundning

1. Create an invoice where rounding applies (see Test 18)
2. Download the PDF
3. Verify the rounding line appears in the totals section

**Pass criteria:** Rounding visible in PDF.

## Test 25: PDF for different statuses

1. Download PDF for a draft invoice — should work (preview before sending)
2. Download PDF for a sent invoice — should work
3. Download PDF for a paid invoice — should work (for records)
4. Download PDF for a cancelled invoice — should work but ideally shows "MAKULERAD" watermark (nice-to-have, not required)

**Pass criteria:** PDF downloadable for all non-deleted statuses.

---

## Edge cases to verify (Fas 2-3)

- Creating an invoice with only one line item
- Creating an invoice with 0% VAT (export, exempt sales)
- Invoice with very large amounts (e.g. 10 000 000 kr)
- Invoice with quantity = 0.5 (half units — should calculate correctly)
- Invoice with 100% discount on a line — line total should be 0
- Trying to finalize an already-sent invoice — should fail
- Trying to pay an already-paid invoice — should fail
- Trying to cancel a paid invoice — should fail
- Trying to delete a sent invoice — should fail
- Two users creating invoices simultaneously — invoice numbers should not collide (atomic increment)
