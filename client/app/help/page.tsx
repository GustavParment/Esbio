"use client";

import { useState } from "react";
import DashboardLayout from "@/components/layout/DashboardLayout";

const sections = [
  {
    id: "getting-started",
    title: "Kom igång",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
    ),
    content: `
### Välkommen till Esbio!

Esbio är ett svenskt bokföringssystem byggt för småföretag. Här är en snabb översikt:

1. **Skapa ett företag** — Gå till Företag och skapa ditt bolag med organisationsnummer
2. **Kontoplan** — BAS-kontoplanen laddas automatiskt
3. **Börja bokföra** — Skapa verifikat under Verifikat → Nytt verifikat
4. **Fakturera** — Lägg upp kunder och skicka fakturor
5. **Rapporter** — Se resultaträkning, balansräkning och momsrapport

**Tips:** Använd Ester AI för att snabbt skapa verifikat med naturligt språk!
    `,
  },
  {
    id: "bookkeeping",
    title: "Bokföring",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
    ),
    content: `
### Grunderna i bokföring

Varje ekonomisk händelse bokförs som ett **verifikat** med minst två rader — debet och kredit måste alltid vara lika.

#### Vanliga konteringar

| Händelse | Debet | Kredit |
|----------|-------|--------|
| Försäljning tjänst (25% moms) | 1930 Bank | 3010 Försäljning + 2610 Utg. moms |
| Inköp material | 4010 Material | 1930 Bank |
| Lokalhyra | 5010 Lokalhyra | 1930 Bank |
| Löner | 7010 Löner | 1930 Bank |
| Konsultarvode | 6550 Konsultarvoden | 1930 Bank |

#### Moms

- **25%** — de flesta varor och tjänster
- **12%** — livsmedel, hotell, restaurang
- **6%** — böcker, tidningar, kultur, kollektivtrafik
- **0%** — sjukvård, utbildning, export

#### Bokslut

Vid räkenskapsårets slut behöver du:
1. Stämma av alla konton mot kontoutdrag
2. Göra periodiseringar (förutbetalda kostnader etc.)
3. Exportera SIE-fil (Rapporter → SIE-export) till din revisor
    `,
  },
  {
    id: "ester",
    title: "Ester AI",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
    ),
    content: `
### Ester AI — din bokföringsassistent

Ester kan hjälpa dig med bokföring genom att förstå naturligt språk.

#### Exempel på vad du kan säga

- *"Bokför försäljning på 10 000 kr exkl moms"*
- *"Lokalhyra 8500 kr för mars"*
- *"Rätta verifikat 5, det ska vara konto 6550 istället för 6010"*
- *"Visa alla verifikat för februari"*
- *"Skicka faktura 3 via mail"*
- *"Hur bokför jag en representation?"*

#### Schemalagda uppgifter

Ester kan schemalägga återkommande bokföring:
- *"Schemalägg lokalhyra 8500 kr den 25:e varje månad"*

Hanteras under Ester AI → Schemalagda uppgifter.

#### Tips

- Var specifik med belopp och moms
- Ester använder dagens datum om du inte anger annat
- Du kan be Ester förklara en kontering
    `,
  },
  {
    id: "receipts",
    title: "Kvittoskanning",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
    ),
    content: `
### Kvittoskanning

Skanna kvitton och låt AI:n föreslå bokföring automatiskt.

#### Så här gör du

1. Gå till **Verifikat → Skanna kvitto**
2. Ta en bild eller ladda upp en bild på kvittot
3. AI:n läser av belopp, datum, butik och moms
4. Granska det föreslagna verifikatet
5. Godkänn och spara

#### Tips för bästa resultat

- Se till att kvittot är väl belyst och i fokus
- Hela kvittot ska synas i bilden
- Digitala kvitton (PDF/skärmdump) fungerar också
- Dubbelkolla alltid momsbeloppet
    `,
  },
  {
    id: "invoicing",
    title: "Fakturering",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="3" width="20" height="18" rx="2"/><line x1="8" y1="7" x2="16" y2="7"/><line x1="8" y1="11" x2="16" y2="11"/><line x1="8" y1="15" x2="12" y2="15"/></svg>
    ),
    content: `
### Fakturering

Skapa och skicka fakturor med automatisk bokföring.

#### Steg för steg

1. **Lägg upp kund** — Kunder → Ny kund (namn, org.nr, adress, e-post)
2. **Fakturainställningar** — Fakturor → Inställningar (bankgiro, plusgiro, betalningsvillkor)
3. **Skapa faktura** — Fakturor → Ny faktura, välj kund och lägg till rader
4. **Skicka** — Klicka "Skicka faktura" för att finalisera (skapar bokföringsverifikat automatiskt)
5. **E-posta** — Klicka "Skicka via e-post" för att maila PDF:en till kunden
6. **Markera betald** — När betalningen kommer in, markera fakturan som betald

#### Automatisk bokföring

När du finaliserar en faktura skapas ett verifikat automatiskt:
- **Debet 1510** (Kundfordringar) — totalbeloppet
- **Kredit 3010** (Försäljning) — nettobeloppet
- **Kredit 2610** (Utgående moms) — momsbeloppet

När fakturan markeras som betald:
- **Debet 1930** (Bank) — totalbeloppet
- **Kredit 1510** (Kundfordringar) — totalbeloppet

#### OCR och betalning

Varje faktura får ett unikt OCR-nummer (Luhn-validerat) som kunden anger vid betalning.
    `,
  },
  {
    id: "reports",
    title: "Rapporter",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
    ),
    content: `
### Rapporter

#### Resultaträkning
Visar intäkter minus kostnader för en vald period. Bra för att se om företaget går med vinst.

#### Balansräkning
Visar tillgångar, skulder och eget kapital vid ett visst datum. Tillgångar = Skulder + Eget kapital.

#### Momsrapport
Sammanställer utgående och ingående moms för en period. Använd siffrorna för att fylla i momsdeklarationen på Skatteverkets hemsida.

#### SIE-export
Exporterar all bokföring i SIE4-format — standardformatet som alla svenska bokföringsprogram förstår. Skicka filen till din revisor vid bokslut.
    `,
  },
  {
    id: "tips",
    title: "Bokföringstips",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 18h6"/><path d="M10 22h4"/><path d="M15.09 14c.18-.98.65-1.74 1.41-2.5A4.65 4.65 0 0 0 18 8 6 6 0 0 0 6 8c0 1 .23 2.23 1.5 3.5A4.61 4.61 0 0 1 8.91 14"/></svg>
    ),
    content: `
### Bokföringstips för småföretag

#### Löpande bokföring
- Bokför löpande — vänta inte till månadens slut
- Spara alltid underlag (kvitton, fakturor) digitalt
- Stäm av bankkontot mot bokföringen regelbundet

#### Moms
- Momsdeklaration görs kvartalsvis eller månadsvis beroende på omsättning
- Under 40 miljoner i omsättning → kvartalsvis (de flesta småföretag)
- Fyll i ruta 05-08 (utgående moms) och ruta 48 (ingående moms) på Skatteverkets blankett

#### Vanliga misstag
- **Blanda ihop privat och företag** — ha separata konton
- **Glömma moms** — bokför alltid momsen separat
- **Felaktig periodisering** — bokför kostnader i rätt månad
- **Tappa kvitton** — fotografera direkt med kvittoskanning

#### Viktiga datum
- **12:e varje månad** — momsdeklaration (om månadsredovisning)
- **12:e i månaden efter kvartalets slut** — kvartalsvis momsdeklaration
- **1 juli** — inkomstdeklaration för aktiebolag (kalenderår)
- **2 maj** — inkomstdeklaration för enskild firma

#### Användbara länkar
- [Skatteverket — Moms](https://www.skatteverket.se/foretag/moms)
- [Skatteverket — Bokföring](https://www.skatteverket.se/foretag/bokforing)
- [BAS-kontoplanen](https://www.bas.se)
    `,
  },
];

export default function HelpPage() {
  const [activeSection, setActiveSection] = useState("getting-started");

  const active = sections.find((s) => s.id === activeSection) || sections[0];

  return (
    <DashboardLayout>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Hjälp</h1>
        <p className="text-gray-600 mt-2">Guider och tips för att komma igång med Esbio</p>
      </div>

      {/* Section cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-8">
        {sections.map((section) => (
          <button
            key={section.id}
            onClick={() => setActiveSection(section.id)}
            className={`p-4 rounded-xl border text-left transition-all ${
              activeSection === section.id
                ? "border-green-500 bg-green-50 shadow-sm"
                : "border-gray-200 bg-white hover:border-gray-300 hover:shadow-sm"
            }`}
          >
            <div className={`mb-2 ${activeSection === section.id ? "text-green-600" : "text-gray-400"}`}>
              {section.icon}
            </div>
            <p className={`text-sm font-medium ${activeSection === section.id ? "text-green-900" : "text-gray-900"}`}>
              {section.title}
            </p>
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 md:p-8">
        <div className="prose prose-gray max-w-none prose-headings:text-gray-900 prose-p:text-gray-700 prose-li:text-gray-700 prose-strong:text-gray-900 prose-a:text-green-600 prose-a:no-underline hover:prose-a:underline prose-table:text-sm prose-th:text-left prose-th:py-2 prose-th:px-3 prose-th:bg-gray-50 prose-td:py-2 prose-td:px-3 prose-td:border-t">
          <div dangerouslySetInnerHTML={{ __html: markdownToHtml(active.content) }} />
        </div>
      </div>

      {/* Contact support */}
      <div className="mt-6 bg-gray-50 rounded-xl border border-gray-200 p-6 text-center">
        <p className="text-gray-700 font-medium mb-1">Hittar du inte svaret?</p>
        <p className="text-gray-500 text-sm mb-3">Kontakta oss via supportformuläret i Inställningar eller maila direkt.</p>
        <a href="mailto:info@esbio.se" className="text-green-600 hover:text-green-700 font-medium text-sm">
          info@esbio.se
        </a>
      </div>
    </DashboardLayout>
  );
}

// Simple markdown-to-HTML converter for the help content
function markdownToHtml(md: string): string {
  const lines = md.trim().split('\n');
  const html: string[] = [];
  let inList = false;
  let inTable = false;
  let tableHeader = false;

  for (const rawLine of lines) {
    let line = rawLine.trim();
    if (!line) {
      if (inList) { html.push('</ul>'); inList = false; }
      if (inTable) { html.push('</tbody></table>'); inTable = false; }
      continue;
    }

    // Inline formatting
    line = line.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
    line = line.replace(/\*(.+?)\*/g, '<em>$1</em>');
    line = line.replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>');

    if (line.startsWith('### ')) {
      if (inList) { html.push('</ul>'); inList = false; }
      html.push(`<h3>${line.slice(4)}</h3>`);
    } else if (line.startsWith('#### ')) {
      if (inList) { html.push('</ul>'); inList = false; }
      html.push(`<h4>${line.slice(5)}</h4>`);
    } else if (line.startsWith('| ') && line.endsWith(' |')) {
      const cells = line.split('|').filter(c => c.trim()).map(c => c.trim());
      if (cells.every(c => /^-+$/.test(c))) {
        tableHeader = false;
        continue;
      }
      if (!inTable) {
        html.push('<table><thead>');
        html.push('<tr>' + cells.map(c => `<th>${c}</th>`).join('') + '</tr>');
        html.push('</thead><tbody>');
        inTable = true;
        tableHeader = true;
      } else {
        html.push('<tr>' + cells.map(c => `<td>${c}</td>`).join('') + '</tr>');
      }
    } else if (line.startsWith('- ')) {
      if (!inList) { html.push('<ul>'); inList = true; }
      html.push(`<li>${line.slice(2)}</li>`);
    } else if (/^\d+\. /.test(line)) {
      if (!inList) { html.push('<ul>'); inList = true; }
      html.push(`<li>${line.replace(/^\d+\. /, '')}</li>`);
    } else {
      if (inList) { html.push('</ul>'); inList = false; }
      html.push(`<p>${line}</p>`);
    }
  }

  if (inList) html.push('</ul>');
  if (inTable) html.push('</tbody></table>');

  return html.join('\n');
}
