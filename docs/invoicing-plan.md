# Fakturering (Invoicing) — Implementation Plan

**Branch:** `feature/invoicing`
**Status:** Fas 1 complete, Fas 2 in progress
**Date:** 2026-04-12

### Fas 1 status (complete)
- DB migrations: `customers` + `invoice_settings` tables created
- Backend: customer CRUD (6 endpoints) + invoice settings (2 endpoints), all wired
- Frontend: customer list/create/detail pages, invoice settings page, sidebar nav updated
- `go build` + `go vet` + `npx tsc --noEmit` all clean
- See [invoicing-test-guide.md](./invoicing-test-guide.md) for manual test steps

---

## Översikt

Faktureringsmodul för Esbio som följer befintlig arkitektur (Go/Gin backend, Next.js frontend, PostgreSQL). Genererar automatiskt bokföringsverifikat vid fakturering och betalning.

---

## DB-schema (4 nya tabeller)

### customers
```sql
CREATE TABLE IF NOT EXISTS customers (
    customer_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    name VARCHAR(200) NOT NULL,
    org_number VARCHAR(20),
    vat_number VARCHAR(30),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    postal_code VARCHAR(10),
    city VARCHAR(100),
    country VARCHAR(100) DEFAULT 'Sverige',
    email VARCHAR(255),
    phone VARCHAR(50),
    payment_terms_days INT NOT NULL DEFAULT 30,
    default_revenue_account INT DEFAULT 3010,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
);
```

### invoice_settings (per företag)
```sql
CREATE TABLE IF NOT EXISTS invoice_settings (
    settings_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL UNIQUE,
    bankgiro VARCHAR(20),
    plusgiro VARCHAR(20),
    swish VARCHAR(20),
    iban VARCHAR(34),
    bic VARCHAR(11),
    f_skatt_text VARCHAR(500) DEFAULT 'Godkänd för F-skatt',
    default_payment_terms_days INT NOT NULL DEFAULT 30,
    next_invoice_number INT NOT NULL DEFAULT 1,
    invoice_prefix VARCHAR(10) DEFAULT '',
    default_revenue_account INT DEFAULT 3010,
    default_payment_account INT DEFAULT 1930,
    footer_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
);
```

### invoices
```sql
CREATE TABLE IF NOT EXISTS invoices (
    invoice_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    customer_id INT NOT NULL,
    invoice_number INT NOT NULL,
    ocr_number VARCHAR(25) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'sent', 'paid', 'overdue', 'cancelled')),
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    payment_terms_days INT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'SEK',
    subtotal DECIMAL(15, 2) NOT NULL DEFAULT 0,
    vat_total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    rounding DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    amount_paid DECIMAL(15, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    your_reference VARCHAR(255),
    our_reference VARCHAR(255),
    bankgiro VARCHAR(20),
    plusgiro VARCHAR(20),
    f_skatt_text VARCHAR(500),
    sales_voucher_id INT,
    payment_voucher_id INT,
    payment_account INT DEFAULT 1930,
    revenue_account INT DEFAULT 3010,
    paid_at TIMESTAMP,
    sent_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE RESTRICT,
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id) ON DELETE RESTRICT,
    FOREIGN KEY (sales_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL,
    FOREIGN KEY (payment_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_invoices_company_number ON invoices(company_id, invoice_number);
```

### invoice_lines
```sql
CREATE TABLE IF NOT EXISTS invoice_lines (
    line_id SERIAL PRIMARY KEY,
    invoice_id INT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    description TEXT NOT NULL,
    quantity DECIMAL(15, 4) NOT NULL DEFAULT 1,
    unit VARCHAR(20) DEFAULT 'st',
    unit_price DECIMAL(15, 2) NOT NULL,
    discount_percent DECIMAL(5, 2) NOT NULL DEFAULT 0,
    vat_rate INT NOT NULL DEFAULT 25 CHECK (vat_rate IN (0, 6, 12, 25)),
    line_total DECIMAL(15, 2) NOT NULL,
    vat_amount DECIMAL(15, 2) NOT NULL,
    account_no INT NOT NULL DEFAULT 3010,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE
);
```

---

## Auto-bokföring

### Vid fakturering (draft → sent)

Skapar ett verifikat automatiskt:

| Konto | Debet | Kredit | Beskrivning |
|-------|-------|--------|-------------|
| 1510 Kundfordringar | totalbelopp inkl moms | | |
| 3010 Försäljning* | | belopp exkl moms | *per rad, kan vara 3040 etc |
| 2610 Utg. moms 25% | | momsbelopp 25% | |
| 2611 Utg. moms 12% | | momsbelopp 12% | |
| 2612 Utg. moms 6% | | momsbelopp 6% | |
| 3740 Öresavrundning | | ev. avrundning | om != 0 |

### Vid betalning (sent → paid)

| Konto | Debet | Kredit |
|-------|-------|--------|
| 1930 Bank (eller valt konto) | totalbelopp | |
| 1510 Kundfordringar | | totalbelopp |

### Vid makulering (cancel)

Skapar ett rättelseverifikat via befintlig `CreateCorrectionVoucher`.

---

## OCR-nummer (svensk standard, Luhn)

```go
func GenerateOCRNumber(invoiceNumber int) string {
    numStr := strconv.Itoa(invoiceNumber)
    sum := 0
    double := true
    for i := len(numStr) - 1; i >= 0; i-- {
        digit := int(numStr[i] - '0')
        if double {
            digit *= 2
            if digit > 9 { digit -= 9 }
        }
        sum += digit
        double = !double
    }
    checkDigit := (10 - (sum % 10)) % 10
    return numStr + strconv.Itoa(checkDigit)
}
```

---

## Öresavrundning

Sverige har inga 50-öresMynt — totaler avrundas till hel krona:

```go
total = subtotal + vatTotal
rounded = total.Round(0)  // nearest whole krona
rounding = rounded.Sub(total)
```

Differensen bokförs på konto 3740.

---

## API-endpoints

### Kunder
| Method | Endpoint | Beskrivning |
|--------|----------|-------------|
| POST | `/customers` | Skapa kund |
| GET | `/customers` | Lista kunder |
| GET | `/customers/search?q=` | Sök kunder |
| GET | `/customers/:id` | Hämta kund |
| PUT | `/customers/:id` | Uppdatera kund |
| DELETE | `/customers/:id` | Radera kund |

### Fakturor
| Method | Endpoint | Beskrivning |
|--------|----------|-------------|
| POST | `/invoices` | Skapa faktura (utkast) |
| GET | `/invoices` | Lista fakturor (?status=, ?customer_id=) |
| GET | `/invoices/:id` | Hämta faktura med rader + kund |
| PUT | `/invoices/:id` | Uppdatera utkast |
| DELETE | `/invoices/:id` | Radera utkast |
| POST | `/invoices/:id/finalize` | Skicka faktura (skapar verifikat) |
| POST | `/invoices/:id/pay` | Markera betald (skapar betalningsverifikat) |
| POST | `/invoices/:id/cancel` | Makulera (skapar rättelseverifikat) |
| GET | `/invoices/:id/pdf` | Ladda ner PDF |
| GET | `/invoices/settings` | Fakturainställningar |
| PUT | `/invoices/settings` | Uppdatera inställningar |

---

## Backend-filer (Go)

| Fil | Innehåll |
|-----|----------|
| `domain/models.go` | Customer, Invoice, InvoiceLine, InvoiceSettings structs |
| `repository/customer_repository.go` | CRUD mot customers-tabell |
| `repository/invoice_repository.go` | CRUD mot invoices |
| `repository/invoice_line_repository.go` | CRUD mot invoice_lines |
| `repository/invoice_settings_repository.go` | Get/Upsert + atomär nummergenerering |
| `service/customer_service.go` | Validering + pass-through |
| `service/invoice_settings_service.go` | Get/upsert med default-skapande |
| `service/invoice_service.go` | **Kärnlogik**: beräkningar, OCR, auto-verifikat, statusflöden |
| `handlers/customer_handler.go` | HTTP-handlers |
| `handlers/invoice_handler.go` | HTTP-handlers |
| `handlers/invoice_pdf_handler.go` | PDF-generering med fpdf |
| `routes/routes.go` | Nya kundgrupp + fakturagrupp |
| `main.go` | DI-wiring |

---

## Frontend-filer (Next.js)

| Fil | Innehåll |
|-----|----------|
| `types/index.ts` | Customer, Invoice, InvoiceLine, InvoiceSettings interfaces |
| `lib/api/customers.ts` | API-klient |
| `lib/api/invoices.ts` | API-klient |
| `app/customers/page.tsx` | Kundlista med sök |
| `app/customers/new/page.tsx` | Skapa kund-formulär |
| `app/customers/[id]/page.tsx` | Kunddetalj + fakturahistorik |
| `app/invoices/page.tsx` | Fakturalista med statusflikar |
| `app/invoices/new/page.tsx` | Skapa faktura med radeditor |
| `app/invoices/[id]/page.tsx` | Fakturavy med åtgärdsknappar |
| `app/invoices/[id]/edit/page.tsx` | Redigera utkast |
| `app/invoices/settings/page.tsx` | Fakturainställningar |
| `components/layout/Sidebar.tsx` | Lägg till Fakturor + Kunder i navigering |

---

## Implementationsfaser

### Fas 1: Kundregister + Fakturainställningar (1-2 dagar)
- Migrations 010 + 011
- Customer + InvoiceSettings: models, repos, services, handlers
- Frontend: kundlista, skapa/redigera kund, inställningssida
- **Shippable**: användare kan hantera kunder och konfigurera bankgiro/plusgiro

### Fas 2: Faktura CRUD + Auto-bokföring (2-3 dagar)
- Migrations 012 + 013
- Invoice + InvoiceLine: models, repos
- InvoiceService med CreateInvoice, FinalizeInvoice, MarkAsPaid, CancelInvoice
- OCR-generering, öresavrundning, verifikatskapande
- Frontend: fakturalista, skapa faktura med radeditor, detaljvy med åtgärdsknappar
- **Shippable**: fullständig fakturering med automatisk bokföring

### Fas 3: Faktura-PDF (1 dag)
- `invoice_pdf_handler.go` med fpdf
- Svensk fakturalayout: företagsinfo, kundinfo, rader, momssummering, betalningsinfo, OCR, F-skatt
- **Shippable**: professionella PDF-fakturor

### Fas 4: Polish + Förfallna fakturor (0.5 dagar)
- `CheckOverdueInvoices` i schedulern (befintlig cron-tjänst)
- Dashboard-widget: utestående fakturor, förfallna
- **Shippable**: automatisk förfallodetektering, dashboard-integration

### Fas 5 (framtid): E-postutskick
- SMTP-integration eller extern tjänst (Resend, SendGrid)
- "Skicka via e-post"-knapp på fakturadetaljen
- Inte del av initial implementation

---

## BAS-konton som berörs

| Konto | Namn | Roll |
|-------|------|------|
| 1510 | Kundfordringar | Debet vid fakturering, kredit vid betalning |
| 1910/1920/1930 | Kassa/PlusGiro/Bank | Debet vid betalning |
| 2610 | Utgående moms 25% | Kredit vid fakturering |
| 2611 | Utgående moms 12% | Kredit vid fakturering |
| 2612 | Utgående moms 6% | Kredit vid fakturering |
| 3010+ | Försäljningsintäkter | Kredit vid fakturering (per rad) |
| 3740 | Öresavrundning | Avrundningsdifferens |

---

## Designbeslut

1. **Verifikat skapas vid finalisering**, inte vid utkast — matchar svensk bokföringslag (den ekonomiska händelsen är fakturautskicket)
2. **Fakturanummer**: atomär sekvens i `invoice_settings` — trådsäkert, inga hål, admins kan ställa startnummer
3. **Snapshot av betalningsuppgifter**: bankgiro/plusgiro/F-skatt kopieras till fakturaraden vid skapande — om inställningar ändras senare påverkas inte befintliga fakturor
4. **Intäktskonto per rad**: varje fakturarad kan ha eget konto (3010, 3040, etc.) — flexibelt för tjänste- vs varuförsäljning
5. **Pengar**: `decimal.Decimal` (Go) / `MoneyString` (TS) genomgående, precis som resten av systemet
