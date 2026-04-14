-- Esbio Bookkeeping Database Schema
-- This file combines all migrations for initial database setup

-- Migration 001: Create users table
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    company_name VARCHAR(200),
    org_number VARCHAR(20),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'Bookkeeper',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- Migration 002: Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    account_no INT PRIMARY KEY,
    account_name VARCHAR(255) NOT NULL,
    account_group INT NOT NULL CHECK (account_group >= 1 AND account_group <= 8),
    tax_standard VARCHAR(50),
    type VARCHAR(10) NOT NULL CHECK (type IN ('P&L', 'BS')),
    standard_side VARCHAR(10) NOT NULL CHECK (standard_side IN ('Debit', 'Credit')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounts_group ON accounts(account_group);
CREATE INDEX idx_accounts_type ON accounts(type);

-- Migration 003: Create companies table
CREATE TABLE IF NOT EXISTS companies (
    company_id SERIAL PRIMARY KEY,
    company_name VARCHAR(200) NOT NULL,
    org_number VARCHAR(20),
    plan VARCHAR(20) NOT NULL DEFAULT 'free',
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

CREATE INDEX idx_companies_created_by ON companies(created_by);

-- Migration 004: Create vouchers table
CREATE TABLE IF NOT EXISTS vouchers (
    voucher_id SERIAL PRIMARY KEY,
    voucher_number INTEGER NOT NULL,
    date DATE NOT NULL,
    description TEXT NOT NULL,
    reference VARCHAR(255),
    total_amount DECIMAL(15, 2) NOT NULL,
    period VARCHAR(7) NOT NULL,
    created_by INT NOT NULL,
    company_id INT NOT NULL,
    corrects_voucher_id INT NULL,
    corrected_by_voucher_id INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE RESTRICT,
    FOREIGN KEY (corrects_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL,
    FOREIGN KEY (corrected_by_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL
);

CREATE INDEX idx_vouchers_period ON vouchers(period);
CREATE INDEX idx_vouchers_created_by ON vouchers(created_by);
CREATE INDEX idx_vouchers_company ON vouchers(company_id);
CREATE INDEX idx_vouchers_date ON vouchers(date);
CREATE INDEX idx_vouchers_number ON vouchers(voucher_number);
ALTER TABLE vouchers ADD CONSTRAINT vouchers_company_number_unique UNIQUE (company_id, voucher_number);

-- Migration 004: Create line_items table
CREATE TABLE IF NOT EXISTS line_items (
    line_id SERIAL PRIMARY KEY,
    voucher_id INT NOT NULL,
    account_no INT NOT NULL,
    debit_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    credit_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    tax_code INT,
    project_id INT,
    cost_center_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (voucher_id) REFERENCES vouchers(voucher_id) ON DELETE CASCADE,
    FOREIGN KEY (account_no) REFERENCES accounts(account_no) ON DELETE RESTRICT,
    CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (credit_amount > 0 AND debit_amount = 0)
    )
);

CREATE INDEX idx_line_items_voucher ON line_items(voucher_id);
CREATE INDEX idx_line_items_account ON line_items(account_no);

-- Insert default users
-- Password for both users is: Password123
INSERT INTO users (first_name, last_name, email, password_hash, role) VALUES
('Admin', 'Esbio', 'admin@esbio.se', '$2a$10$Cs8h5LLIFMsFN3rWiR0zp.CCFrkzc1.3Kn6IY5QLNVqjx6sx6Srn6', 'Admin'),
('Test', 'User', 'test@gmail.com', '$2a$10$Cs8h5LLIFMsFN3rWiR0zp.CCFrkzc1.3Kn6IY5QLNVqjx6sx6Srn6', 'Bookkeeper');

-- Insert standard Swedish BAS accounts
INSERT INTO accounts (account_no, account_name, account_group, tax_standard, type, standard_side) VALUES
-- KLASS 1: TILLGÅNGAR (Assets)
-- 10 Immateriella anläggningstillgångar
(1010, 'Utvecklingsutgifter', 1, '0%', 'BS', 'Debit'),
(1020, 'Koncessioner', 1, '0%', 'BS', 'Debit'),
(1030, 'Patent', 1, '0%', 'BS', 'Debit'),
(1040, 'Licenser', 1, '0%', 'BS', 'Debit'),
(1050, 'Varumärken', 1, '0%', 'BS', 'Debit'),
(1060, 'Hyresrätter', 1, '0%', 'BS', 'Debit'),
(1070, 'Goodwill', 1, '0%', 'BS', 'Debit'),
(1080, 'Pågående projekt immateriella', 1, '0%', 'BS', 'Debit'),
(1090, 'Förskott immateriella tillgångar', 1, '0%', 'BS', 'Debit'),

-- 11 Byggnader och mark
(1110, 'Byggnader', 1, '0%', 'BS', 'Debit'),
(1120, 'Byggnadsinventarier', 1, '0%', 'BS', 'Debit'),
(1130, 'Mark', 1, '0%', 'BS', 'Debit'),
(1140, 'Markanläggningar', 1, '0%', 'BS', 'Debit'),
(1150, 'Markinventarier', 1, '0%', 'BS', 'Debit'),
(1180, 'Pågående nyanläggningar fastigheter', 1, '0%', 'BS', 'Debit'),
(1190, 'Förskott för byggnader och mark', 1, '0%', 'BS', 'Debit'),

-- 12 Maskiner och inventarier
(1210, 'Maskiner och andra tekniska anläggningar', 1, '0%', 'BS', 'Debit'),
(1220, 'Inventarier och verktyg', 1, '0%', 'BS', 'Debit'),
(1230, 'Installationer', 1, '0%', 'BS', 'Debit'),
(1240, 'Bilar och andra transportmedel', 1, '0%', 'BS', 'Debit'),
(1250, 'Datorer', 1, '0%', 'BS', 'Debit'),
(1260, 'Leasade tillgångar', 1, '0%', 'BS', 'Debit'),
(1280, 'Pågående nyanläggningar inventarier', 1, '0%', 'BS', 'Debit'),
(1290, 'Förskott för maskiner och inventarier', 1, '0%', 'BS', 'Debit'),

-- 13 Finansiella anläggningstillgångar
(1310, 'Andelar i koncernföretag', 1, '0%', 'BS', 'Debit'),
(1320, 'Långfristiga fordringar hos koncernföretag', 1, '0%', 'BS', 'Debit'),
(1330, 'Andelar i intresseföretag', 1, '0%', 'BS', 'Debit'),
(1340, 'Långfristiga fordringar hos intresseföretag', 1, '0%', 'BS', 'Debit'),
(1350, 'Andelar och värdepapper i andra företag', 1, '0%', 'BS', 'Debit'),
(1360, 'Långfristiga fordringar hos andra företag', 1, '0%', 'BS', 'Debit'),
(1380, 'Andra långfristiga fordringar', 1, '0%', 'BS', 'Debit'),

-- 14 Lager
(1410, 'Lager av råvaror', 1, '0%', 'BS', 'Debit'),
(1420, 'Lager av tillsatsmaterial', 1, '0%', 'BS', 'Debit'),
(1430, 'Lager av halvfabrikat', 1, '0%', 'BS', 'Debit'),
(1440, 'Lager av produkter i arbete', 1, '0%', 'BS', 'Debit'),
(1450, 'Lager av färdiga varor', 1, '0%', 'BS', 'Debit'),
(1460, 'Lager av handelsvaror', 1, '0%', 'BS', 'Debit'),
(1470, 'Pågående arbeten', 1, '0%', 'BS', 'Debit'),
(1480, 'Förskott till leverantörer', 1, '0%', 'BS', 'Debit'),

-- 15 Kundfordringar
(1510, 'Kundfordringar', 1, '0%', 'BS', 'Debit'),
(1511, 'Kundfordringar hos koncernföretag', 1, '0%', 'BS', 'Debit'),
(1512, 'Kundfordringar hos intresseföretag', 1, '0%', 'BS', 'Debit'),
(1515, 'Osäkra kundfordringar', 1, '0%', 'BS', 'Debit'),
(1519, 'Nedskrivning av kundfordringar', 1, '0%', 'BS', 'Credit'),

-- 16 Övriga kortfristiga fordringar
(1610, 'Fordringar hos anställda', 1, '0%', 'BS', 'Debit'),
(1620, 'Fordringar hos koncernföretag', 1, '0%', 'BS', 'Debit'),
(1630, 'Fordringar hos intresseföretag', 1, '0%', 'BS', 'Debit'),
(1640, 'Skattefordringar', 1, '0%', 'BS', 'Debit'),
(1650, 'Momsfordran', 1, '0%', 'BS', 'Debit'),
(1680, 'Övriga kortfristiga fordringar', 1, '0%', 'BS', 'Debit'),

-- 17 Förutbetalda kostnader och upplupna intäkter
(1710, 'Förutbetalda hyreskostnader', 1, '0%', 'BS', 'Debit'),
(1720, 'Förutbetalda leasingavgifter', 1, '0%', 'BS', 'Debit'),
(1730, 'Förutbetalda försäkringspremier', 1, '0%', 'BS', 'Debit'),
(1740, 'Förutbetalda räntekostnader', 1, '0%', 'BS', 'Debit'),
(1750, 'Upplupna hyresintäkter', 1, '0%', 'BS', 'Debit'),
(1760, 'Upplupna ränteintäkter', 1, '0%', 'BS', 'Debit'),
(1790, 'Övriga förutbetalda kostnader och upplupna intäkter', 1, '0%', 'BS', 'Debit'),

-- 18 Kortfristiga placeringar
(1810, 'Andelar i börsnoterade företag', 1, '0%', 'BS', 'Debit'),
(1820, 'Obligationer', 1, '0%', 'BS', 'Debit'),
(1830, 'Konvertibla skuldebrev', 1, '0%', 'BS', 'Debit'),
(1880, 'Övriga kortfristiga placeringar', 1, '0%', 'BS', 'Debit'),
(1890, 'Nedskrivning av kortfristiga placeringar', 1, '0%', 'BS', 'Credit'),

-- 19 Kassa och bank
(1910, 'Kassa', 1, '0%', 'BS', 'Debit'),
(1920, 'PlusGiro', 1, '0%', 'BS', 'Debit'),
(1930, 'Företagskonto / checkkonto', 1, '0%', 'BS', 'Debit'),
(1940, 'Övriga bankkonton', 1, '0%', 'BS', 'Debit'),
(1950, 'Bankcertifikat', 1, '0%', 'BS', 'Debit'),
(1960, 'Koncernkonto', 1, '0%', 'BS', 'Debit'),

-- KLASS 2: EGET KAPITAL OCH SKULDER (Equity and Liabilities)
-- 20 Eget kapital
(2010, 'Eget kapital, enskild firma', 2, '0%', 'BS', 'Credit'),
(2011, 'Egna uttag', 2, '0%', 'BS', 'Debit'),
(2012, 'Egna insättningar', 2, '0%', 'BS', 'Credit'),
(2013, 'Årets resultat', 2, '0%', 'BS', 'Credit'),
(2081, 'Aktiekapital', 2, '0%', 'BS', 'Credit'),
(2082, 'Ej registrerat aktiekapital', 2, '0%', 'BS', 'Credit'),
(2085, 'Uppskrivningsfond', 2, '0%', 'BS', 'Credit'),
(2086, 'Reservfond', 2, '0%', 'BS', 'Credit'),
(2087, 'Insatskapital', 2, '0%', 'BS', 'Credit'),
(2090, 'Fritt eget kapital', 2, '0%', 'BS', 'Credit'),
(2091, 'Balanserad vinst eller förlust', 2, '0%', 'BS', 'Credit'),
(2097, 'Överkursfond', 2, '0%', 'BS', 'Credit'),
(2098, 'Vinst eller förlust från föregående år', 2, '0%', 'BS', 'Credit'),
(2099, 'Årets resultat', 2, '0%', 'BS', 'Credit'),

-- 21 Obeskattade reserver
(2110, 'Periodiseringsfonder', 2, '0%', 'BS', 'Credit'),
(2120, 'Ackumulerade överavskrivningar byggnader', 2, '0%', 'BS', 'Credit'),
(2130, 'Ackumulerade överavskrivningar maskiner', 2, '0%', 'BS', 'Credit'),
(2150, 'Ackumulerade överavskrivningar immateriella', 2, '0%', 'BS', 'Credit'),

-- 22 Avsättningar
(2210, 'Avsättningar för pensioner', 2, '0%', 'BS', 'Credit'),
(2220, 'Avsättningar för garantier', 2, '0%', 'BS', 'Credit'),
(2250, 'Övriga avsättningar för skatter', 2, '0%', 'BS', 'Credit'),
(2290, 'Övriga avsättningar', 2, '0%', 'BS', 'Credit'),

-- 23 Långfristiga skulder
(2310, 'Obligations- och förlagslån', 2, '0%', 'BS', 'Credit'),
(2320, 'Konvertibla lån', 2, '0%', 'BS', 'Credit'),
(2330, 'Checkräkningskredit', 2, '0%', 'BS', 'Credit'),
(2340, 'Byggnadskreditiv', 2, '0%', 'BS', 'Credit'),
(2350, 'Banklån', 2, '0%', 'BS', 'Credit'),
(2360, 'Skulder till koncernföretag', 2, '0%', 'BS', 'Credit'),
(2370, 'Skulder till intresseföretag', 2, '0%', 'BS', 'Credit'),
(2390, 'Övriga långfristiga skulder', 2, '0%', 'BS', 'Credit'),
(2393, 'Lån från aktieägare', 2, '0%', 'BS', 'Credit'),
(2395, 'Villkorade aktieägartillskott', 2, '0%', 'BS', 'Credit'),

-- 24 Kortfristiga skulder till kreditinstitut
(2410, 'Kortfristiga skulder till kreditinstitut', 2, '0%', 'BS', 'Credit'),
(2417, 'Kortfristig del av långfristiga skulder', 2, '0%', 'BS', 'Credit'),
(2420, 'Förskott från kunder', 2, '0%', 'BS', 'Credit'),
(2421, 'Erhållen handpenning', 2, '0%', 'BS', 'Credit'),
(2440, 'Leverantörsskulder', 2, '0%', 'BS', 'Credit'),
(2441, 'Leverantörsskulder till koncernföretag', 2, '0%', 'BS', 'Credit'),
(2442, 'Leverantörsskulder till intresseföretag', 2, '0%', 'BS', 'Credit'),
(2450, 'Växelskulder', 2, '0%', 'BS', 'Credit'),

-- 25 Skatteskulder
(2510, 'Skatteskulder', 2, '0%', 'BS', 'Credit'),
(2514, 'Beräknad särskild löneskatt', 2, '0%', 'BS', 'Credit'),
(2515, 'Beräknad avkastningsskatt', 2, '0%', 'BS', 'Credit'),
(2518, 'Betald F-skatt', 2, '0%', 'BS', 'Debit'),

-- 26 Moms och punktskatter
(2610, 'Utgående moms 25%', 2, '25%', 'BS', 'Credit'),
(2611, 'Utgående moms 25% varor', 2, '25%', 'BS', 'Credit'),
(2612, 'Utgående moms 25% tjänster', 2, '25%', 'BS', 'Credit'),
(2620, 'Utgående moms 12%', 2, '12%', 'BS', 'Credit'),
(2630, 'Utgående moms 6%', 2, '6%', 'BS', 'Credit'),
(2640, 'Ingående moms', 2, '0%', 'BS', 'Debit'),
(2641, 'Debiterad ingående moms', 2, '0%', 'BS', 'Debit'),
(2645, 'Beräknad ingående moms på förvärv', 2, '0%', 'BS', 'Debit'),
(2650, 'Redovisningskonto moms', 2, '0%', 'BS', 'Credit'),

-- 27 Personalens skatter, avgifter och lön
(2710, 'Personalskatt', 2, '0%', 'BS', 'Credit'),
(2730, 'Lagstadgade sociala avgifter', 2, '0%', 'BS', 'Credit'),
(2731, 'Avräkning sociala avgifter', 2, '0%', 'BS', 'Credit'),
(2732, 'Arbetsgivaravgifter', 2, '0%', 'BS', 'Credit'),
(2750, 'Utmätning i lön', 2, '0%', 'BS', 'Credit'),
(2760, 'Semesterlöneskuld', 2, '0%', 'BS', 'Credit'),
(2790, 'Övriga skulder till anställda', 2, '0%', 'BS', 'Credit'),
(2791, 'Löneskulder', 2, '0%', 'BS', 'Credit'),

-- 28 Övriga kortfristiga skulder
(2810, 'Avräkning koncernföretag', 2, '0%', 'BS', 'Credit'),
(2820, 'Skuld till delägare', 2, '0%', 'BS', 'Credit'),
(2840, 'Kortfristiga låneskulder', 2, '0%', 'BS', 'Credit'),
(2890, 'Övriga kortfristiga skulder', 2, '0%', 'BS', 'Credit'),
(2891, 'Skuld för ställda säkerheter', 2, '0%', 'BS', 'Credit'),
(2893, 'Outtagen vinstutdelning', 2, '0%', 'BS', 'Credit'),

-- 29 Upplupna kostnader och förutbetalda intäkter
(2910, 'Upplupna löner', 2, '0%', 'BS', 'Credit'),
(2920, 'Upplupna semesterlöner', 2, '0%', 'BS', 'Credit'),
(2930, 'Upplupna sociala avgifter', 2, '0%', 'BS', 'Credit'),
(2940, 'Upplupna räntekostnader', 2, '0%', 'BS', 'Credit'),
(2950, 'Upplupna lagstadgade sociala avgifter', 2, '0%', 'BS', 'Credit'),
(2960, 'Förutbetalda hyresintäkter', 2, '0%', 'BS', 'Credit'),
(2970, 'Förutbetalda intäkter', 2, '0%', 'BS', 'Credit'),
(2990, 'Övriga upplupna kostnader och förutbetalda intäkter', 2, '0%', 'BS', 'Credit'),

-- KLASS 3: RÖRELSENS INTÄKTER (Operating Revenue)
(3000, 'Försäljning varor och tjänster', 3, '25%', 'P&L', 'Credit'),
(3001, 'Försäljning varor 25% moms', 3, '25%', 'P&L', 'Credit'),
(3002, 'Försäljning varor 12% moms', 3, '12%', 'P&L', 'Credit'),
(3003, 'Försäljning varor 6% moms', 3, '6%', 'P&L', 'Credit'),
(3004, 'Försäljning varor momsfri', 3, '0%', 'P&L', 'Credit'),
(3010, 'Försäljning varor', 3, '25%', 'P&L', 'Credit'),
(3040, 'Försäljning tjänster', 3, '25%', 'P&L', 'Credit'),
(3041, 'Försäljning tjänster 25% moms', 3, '25%', 'P&L', 'Credit'),
(3050, 'Försäljning till koncernföretag', 3, '25%', 'P&L', 'Credit'),
(3060, 'Försäljning till intresseföretag', 3, '25%', 'P&L', 'Credit'),
(3100, 'Försäljning livsmedel', 3, '12%', 'P&L', 'Credit'),
(3200, 'Försäljning övrigt', 3, '25%', 'P&L', 'Credit'),
(3211, 'Försäljning böcker och tidskrifter', 3, '6%', 'P&L', 'Credit'),
(3300, 'EU-försäljning varor', 3, '0%', 'P&L', 'Credit'),
(3305, 'EU-försäljning tjänster', 3, '0%', 'P&L', 'Credit'),
(3400, 'Export varor', 3, '0%', 'P&L', 'Credit'),
(3500, 'Fakturerade kostnader', 3, '25%', 'P&L', 'Credit'),
(3510, 'Fakturerade frakter', 3, '25%', 'P&L', 'Credit'),
(3520, 'Fakturerade emballage', 3, '25%', 'P&L', 'Credit'),
(3540, 'Fakturerade tullkostnader', 3, '0%', 'P&L', 'Credit'),
(3590, 'Övriga fakturerade kostnader', 3, '25%', 'P&L', 'Credit'),
(3600, 'Övriga sidointäkter', 3, '25%', 'P&L', 'Credit'),
(3610, 'Intäkter lagerförsäljning', 3, '25%', 'P&L', 'Credit'),
(3690, 'Övriga försäljningsintäkter', 3, '25%', 'P&L', 'Credit'),
(3700, 'Intäktskorrigeringar', 3, '25%', 'P&L', 'Credit'),
(3730, 'Lämnade rabatter', 3, '25%', 'P&L', 'Debit'),
(3731, 'Lämnade kassarabatter', 3, '25%', 'P&L', 'Debit'),
(3740, 'Öresutjämning', 3, '0%', 'P&L', 'Credit'),
(3900, 'Övriga rörelseintäkter', 3, '0%', 'P&L', 'Credit'),
(3910, 'Hyresintäkter', 3, '0%', 'P&L', 'Credit'),
(3911, 'Hyresintäkter lokaler', 3, '25%', 'P&L', 'Credit'),
(3920, 'Provisionsintäkter', 3, '25%', 'P&L', 'Credit'),
(3950, 'Återvunna kundfordringar', 3, '0%', 'P&L', 'Credit'),
(3960, 'Valutakursvinster', 3, '0%', 'P&L', 'Credit'),
(3970, 'Vinst vid avyttring av anläggningstillgångar', 3, '0%', 'P&L', 'Credit'),
(3980, 'Erhållna bidrag', 3, '0%', 'P&L', 'Credit'),
(3985, 'Erhållna statliga bidrag', 3, '0%', 'P&L', 'Credit'),
(3990, 'Övriga ersättningar och intäkter', 3, '0%', 'P&L', 'Credit'),

-- KLASS 4: MATERIAL- OCH VARUKOSTNADER (Cost of Goods)
(4000, 'Inköp varor och material', 4, '25%', 'P&L', 'Debit'),
(4010, 'Inköp varor', 4, '25%', 'P&L', 'Debit'),
(4011, 'Inköp varor 25% moms', 4, '25%', 'P&L', 'Debit'),
(4012, 'Inköp varor 12% moms', 4, '12%', 'P&L', 'Debit'),
(4013, 'Inköp varor 6% moms', 4, '6%', 'P&L', 'Debit'),
(4040, 'Inköp material', 4, '25%', 'P&L', 'Debit'),
(4050, 'Inköp från koncernföretag', 4, '25%', 'P&L', 'Debit'),
(4060, 'Inköp från intresseföretag', 4, '25%', 'P&L', 'Debit'),
(4200, 'Legoarbeten och underentreprenörer', 4, '25%', 'P&L', 'Debit'),
(4400, 'Förbrukningsmaterial', 4, '25%', 'P&L', 'Debit'),
(4500, 'Övriga inköp', 4, '25%', 'P&L', 'Debit'),
(4510, 'Importkostnader', 4, '0%', 'P&L', 'Debit'),
(4530, 'Tull- och speditionskostnader', 4, '0%', 'P&L', 'Debit'),
(4531, 'Tullkostnader', 4, '0%', 'P&L', 'Debit'),
(4532, 'Speditionskostnader', 4, '25%', 'P&L', 'Debit'),
(4545, 'Direkta fraktkostnader', 4, '25%', 'P&L', 'Debit'),
(4600, 'Legoarbeten', 4, '25%', 'P&L', 'Debit'),
(4700, 'Kostnadskorrigeringar', 4, '25%', 'P&L', 'Credit'),
(4730, 'Erhållna rabatter', 4, '25%', 'P&L', 'Credit'),
(4731, 'Erhållna kassarabatter', 4, '25%', 'P&L', 'Credit'),
(4900, 'Förändring lager', 4, '0%', 'P&L', 'Credit'),
(4910, 'Förändring råvaror', 4, '0%', 'P&L', 'Credit'),
(4920, 'Förändring produkter i arbete', 4, '0%', 'P&L', 'Credit'),
(4930, 'Förändring färdiga varor', 4, '0%', 'P&L', 'Credit'),
(4940, 'Förändring handelsvaror', 4, '0%', 'P&L', 'Credit'),
(4990, 'Övrig lagerförändring', 4, '0%', 'P&L', 'Credit'),

-- KLASS 5: PERSONALKOSTNADER (Personnel Costs)
(5010, 'Löner till tjänstemän', 5, '0%', 'P&L', 'Debit'),
(5020, 'Löner till kollektivanställda', 5, '0%', 'P&L', 'Debit'),
(5030, 'Löner till företagsledning', 5, '0%', 'P&L', 'Debit'),
(5050, 'Sjuklön', 5, '0%', 'P&L', 'Debit'),
(5090, 'Förändring semesterlöneskuld', 5, '0%', 'P&L', 'Debit'),
(5100, 'Arvoden till styrelse', 5, '0%', 'P&L', 'Debit'),
(5190, 'Övriga löner', 5, '0%', 'P&L', 'Debit'),
(5191, 'Bonus', 5, '0%', 'P&L', 'Debit'),
(5192, 'Provision', 5, '0%', 'P&L', 'Debit'),
(5200, 'Egenavgifter', 5, '0%', 'P&L', 'Debit'),
(5210, 'Arbetsgivaravgifter', 5, '0%', 'P&L', 'Debit'),
(5220, 'Sociala avgifter styrelsearvoden', 5, '0%', 'P&L', 'Debit'),
(5250, 'Sociala avgifter', 5, '0%', 'P&L', 'Debit'),
(5290, 'Övriga lagstadgade avgifter', 5, '0%', 'P&L', 'Debit'),
(5300, 'Pensionskostnader', 5, '0%', 'P&L', 'Debit'),
(5400, 'Sjuk- och olycksfallsförsäkringar', 5, '0%', 'P&L', 'Debit'),
(5410, 'Grupplivförsäkring', 5, '0%', 'P&L', 'Debit'),
(5420, 'Gruppsjukförsäkring', 5, '0%', 'P&L', 'Debit'),
(5500, 'Övriga personalkostnader', 5, '25%', 'P&L', 'Debit'),
(5600, 'Utbildning', 5, '25%', 'P&L', 'Debit'),
(5610, 'Utbildning extern', 5, '25%', 'P&L', 'Debit'),
(5620, 'Utbildning intern', 5, '25%', 'P&L', 'Debit'),
(5700, 'Sjuk- och hälsovård', 5, '0%', 'P&L', 'Debit'),
(5710, 'Friskvård', 5, '0%', 'P&L', 'Debit'),
(5800, 'Personalrepresentation', 5, '0%', 'P&L', 'Debit'),
(5810, 'Personalfester', 5, '0%', 'P&L', 'Debit'),
(5820, 'Julbord personal', 5, '0%', 'P&L', 'Debit'),

-- KLASS 6: ÖVRIGA EXTERNA KOSTNADER (Other Operating Expenses)
(6000, 'Övriga externa kostnader', 6, '25%', 'P&L', 'Debit'),
(6010, 'Lokalhyra', 6, '25%', 'P&L', 'Debit'),
(6020, 'El för lokaler', 6, '25%', 'P&L', 'Debit'),
(6030, 'Värme', 6, '25%', 'P&L', 'Debit'),
(6040, 'Vatten och avlopp', 6, '25%', 'P&L', 'Debit'),
(6050, 'Lokalvård', 6, '25%', 'P&L', 'Debit'),
(6060, 'Reparation och underhåll lokaler', 6, '25%', 'P&L', 'Debit'),
(6070, 'Hyra av inventarier', 6, '25%', 'P&L', 'Debit'),
(6090, 'Övriga lokalkostnader', 6, '25%', 'P&L', 'Debit'),
(6100, 'Förbrukningsinventarier', 6, '25%', 'P&L', 'Debit'),
(6110, 'Kontorsmaterial', 6, '25%', 'P&L', 'Debit'),
(6150, 'Trycksaker', 6, '25%', 'P&L', 'Debit'),
(6200, 'Telefonkostnader', 6, '25%', 'P&L', 'Debit'),
(6210, 'Fast telefoni', 6, '25%', 'P&L', 'Debit'),
(6211, 'Mobiltelefon', 6, '25%', 'P&L', 'Debit'),
(6212, 'Bredband', 6, '25%', 'P&L', 'Debit'),
(6230, 'Porto', 6, '0%', 'P&L', 'Debit'),
(6250, 'IT-tjänster', 6, '25%', 'P&L', 'Debit'),
(6251, 'Programvaror', 6, '25%', 'P&L', 'Debit'),
(6300, 'Företagsförsäkringar', 6, '0%', 'P&L', 'Debit'),
(6310, 'Bilförsäkringar', 6, '0%', 'P&L', 'Debit'),
(6320, 'Ansvarsförsäkring', 6, '0%', 'P&L', 'Debit'),
(6340, 'Egendomsförsäkring', 6, '0%', 'P&L', 'Debit'),
(6400, 'Styrelse- och revisionsarvoden', 6, '25%', 'P&L', 'Debit'),
(6410, 'Revisionsarvoden', 6, '25%', 'P&L', 'Debit'),
(6420, 'Redovisningstjänster', 6, '25%', 'P&L', 'Debit'),
(6500, 'Övriga externa tjänster', 6, '25%', 'P&L', 'Debit'),
(6510, 'Juridisk rådgivning', 6, '25%', 'P&L', 'Debit'),
(6520, 'Ekonomisk rådgivning', 6, '25%', 'P&L', 'Debit'),
(6530, 'Personaltjänster', 6, '25%', 'P&L', 'Debit'),
(6540, 'Inkassokostnader', 6, '25%', 'P&L', 'Debit'),
(6550, 'Konsultarvoden', 6, '25%', 'P&L', 'Debit'),
(6560, 'Serviceavgifter', 6, '25%', 'P&L', 'Debit'),
(6570, 'Bankkostnader', 6, '0%', 'P&L', 'Debit'),
(6600, 'Frakter och transporter', 6, '25%', 'P&L', 'Debit'),
(6700, 'Resekostnader', 6, '25%', 'P&L', 'Debit'),
(6710, 'Biljetter', 6, '0%', 'P&L', 'Debit'),
(6720, 'Hotell', 6, '12%', 'P&L', 'Debit'),
(6730, 'Traktamenten', 6, '0%', 'P&L', 'Debit'),
(6790, 'Övriga resekostnader', 6, '25%', 'P&L', 'Debit'),
(6800, 'Bilkostnader', 6, '25%', 'P&L', 'Debit'),
(6810, 'Drivmedel', 6, '25%', 'P&L', 'Debit'),
(6820, 'Reparation och underhåll bilar', 6, '25%', 'P&L', 'Debit'),
(6830, 'Bilskatt', 6, '0%', 'P&L', 'Debit'),
(6840, 'Leasingavgifter bilar', 6, '25%', 'P&L', 'Debit'),
(6850, 'Milersättning', 6, '0%', 'P&L', 'Debit'),
(6900, 'Reklam och PR', 6, '25%', 'P&L', 'Debit'),
(6910, 'Annonsering', 6, '25%', 'P&L', 'Debit'),
(6920, 'Mässor och utställningar', 6, '25%', 'P&L', 'Debit'),
(6970, 'Extern representation', 6, '0%', 'P&L', 'Debit'),
(6971, 'Representation avdragsgill', 6, '0%', 'P&L', 'Debit'),
(6972, 'Representation ej avdragsgill', 6, '0%', 'P&L', 'Debit'),
(6980, 'Sponsring', 6, '0%', 'P&L', 'Debit'),
(6990, 'Övriga externa kostnader', 6, '25%', 'P&L', 'Debit'),
(6991, 'Medlemsavgifter', 6, '0%', 'P&L', 'Debit'),
(6992, 'Tidningar och tidskrifter', 6, '6%', 'P&L', 'Debit'),
(6993, 'Facklitteratur', 6, '6%', 'P&L', 'Debit'),

-- KLASS 7: AVSKRIVNINGAR OCH FINANSIELLA POSTER (Depreciation and Financial Items)
(7010, 'Avskrivningar immateriella tillgångar', 7, '0%', 'P&L', 'Debit'),
(7020, 'Avskrivningar byggnader', 7, '0%', 'P&L', 'Debit'),
(7030, 'Avskrivningar maskiner och inventarier', 7, '0%', 'P&L', 'Debit'),
(7050, 'Avskrivningar hyresrätter', 7, '0%', 'P&L', 'Debit'),
(7060, 'Avskrivningar goodwill', 7, '0%', 'P&L', 'Debit'),
(7070, 'Nedskrivningar immateriella tillgångar', 7, '0%', 'P&L', 'Debit'),
(7080, 'Nedskrivningar materiella tillgångar', 7, '0%', 'P&L', 'Debit'),
(7200, 'Förlust vid avyttring av anläggningstillgångar', 7, '0%', 'P&L', 'Debit'),
(7310, 'Nedskrivningar av andelar i koncernföretag', 7, '0%', 'P&L', 'Debit'),
(7320, 'Nedskrivningar av andelar i intresseföretag', 7, '0%', 'P&L', 'Debit'),
(7400, 'Kundförluster', 7, '0%', 'P&L', 'Debit'),
(7500, 'Övriga rörelsekostnader', 7, '0%', 'P&L', 'Debit'),
(7510, 'Valutakursförluster', 7, '0%', 'P&L', 'Debit'),
(7600, 'Bankkostnader', 7, '0%', 'P&L', 'Debit'),

-- 80 Resultat från andelar
(8010, 'Resultat från andelar i koncernföretag', 8, '0%', 'P&L', 'Credit'),
(8020, 'Resultat från andelar i intresseföretag', 8, '0%', 'P&L', 'Credit'),
(8030, 'Resultat från övriga värdepapper', 8, '0%', 'P&L', 'Credit'),
(8100, 'Utdelningar på aktier', 8, '0%', 'P&L', 'Credit'),

-- 82-84 Ränteintäkter och räntekostnader
(8210, 'Ränteintäkter från koncernföretag', 8, '0%', 'P&L', 'Credit'),
(8220, 'Ränteintäkter från intresseföretag', 8, '0%', 'P&L', 'Credit'),
(8300, 'Ränteintäkter', 8, '0%', 'P&L', 'Credit'),
(8310, 'Ränteintäkter bank', 8, '0%', 'P&L', 'Credit'),
(8311, 'Ränteintäkter kundfordringar', 8, '0%', 'P&L', 'Credit'),
(8314, 'Skattefria ränteintäkter', 8, '0%', 'P&L', 'Credit'),
(8380, 'Utdelning på kortfristiga placeringar', 8, '0%', 'P&L', 'Credit'),
(8390, 'Övriga finansiella intäkter', 8, '0%', 'P&L', 'Credit'),
(8400, 'Räntekostnader', 8, '0%', 'P&L', 'Debit'),
(8410, 'Räntekostnader till kreditinstitut', 8, '0%', 'P&L', 'Debit'),
(8420, 'Räntekostnader till koncernföretag', 8, '0%', 'P&L', 'Debit'),
(8430, 'Räntekostnader till intresseföretag', 8, '0%', 'P&L', 'Debit'),
(8440, 'Räntekostnader leverantörsskulder', 8, '0%', 'P&L', 'Debit'),
(8490, 'Övriga finansiella kostnader', 8, '0%', 'P&L', 'Debit'),

-- 85-88 Bokslutsdispositioner
(8810, 'Förändring av periodiseringsfonder', 8, '0%', 'P&L', 'Debit'),
(8820, 'Förändring av överavskrivningar', 8, '0%', 'P&L', 'Debit'),
(8850, 'Förändring av ersättningsfonder', 8, '0%', 'P&L', 'Debit'),
(8890, 'Övriga bokslutsdispositioner', 8, '0%', 'P&L', 'Debit'),

-- 89 Skatter
(8910, 'Skatt på årets resultat', 8, '0%', 'P&L', 'Debit'),
(8920, 'Skatt på grund av ändrad taxering', 8, '0%', 'P&L', 'Debit'),
(8930, 'Latent skatt', 8, '0%', 'P&L', 'Debit'),
(8990, 'Övriga skatter', 8, '0%', 'P&L', 'Debit'),
(8999, 'Årets resultat', 8, '0%', 'P&L', 'Credit')
ON CONFLICT (account_no) DO NOTHING;

-- Migration 005: Create scheduled_tasks table
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    task_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    company_id INT NOT NULL,
    description TEXT NOT NULL,
    prompt TEXT NOT NULL,
    template_voucher_id INT,
    day_of_month INT NOT NULL CHECK (day_of_month >= 1 AND day_of_month <= 31),
    active BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (template_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL
);

CREATE INDEX idx_scheduled_tasks_user ON scheduled_tasks(user_id);
CREATE INDEX idx_scheduled_tasks_company ON scheduled_tasks(company_id);
CREATE INDEX idx_scheduled_tasks_next_run ON scheduled_tasks(next_run_at);
CREATE INDEX idx_scheduled_tasks_active ON scheduled_tasks(active);

-- Migration 006: Create agent_messages table
CREATE TABLE IF NOT EXISTS agent_messages (
    message_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    company_id INT NOT NULL,
    conversation_id VARCHAR(36) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
);

CREATE INDEX idx_agent_messages_user ON agent_messages(user_id);
CREATE INDEX idx_agent_messages_company ON agent_messages(company_id);
CREATE INDEX idx_agent_messages_conversation ON agent_messages(conversation_id);

-- Migration 007: Create agent_usage table
CREATE TABLE IF NOT EXISTS agent_usage (
    usage_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    company_id INT NOT NULL,
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
);

CREATE INDEX idx_agent_usage_user ON agent_usage(user_id);
CREATE INDEX idx_agent_usage_company ON agent_usage(company_id);
CREATE INDEX idx_agent_usage_created ON agent_usage(created_at);

-- Migration 008: Create bank_connections table (Tink Open Banking)
CREATE TABLE IF NOT EXISTS bank_connections (
    connection_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    created_by INT NOT NULL,
    tink_user_id VARCHAR(255) NOT NULL,
    tink_credentials_id VARCHAR(255),
    bank_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expiring_soon', 'expired', 'revoked')),
    access_token_encrypted TEXT,
    consent_expires_at TIMESTAMP NOT NULL,
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

CREATE INDEX idx_bank_connections_company ON bank_connections(company_id);
CREATE INDEX idx_bank_connections_status ON bank_connections(status);

-- Migration 009: Create bank_accounts table
CREATE TABLE IF NOT EXISTS bank_accounts (
    bank_account_id SERIAL PRIMARY KEY,
    connection_id INT NOT NULL,
    company_id INT NOT NULL,
    tink_account_id VARCHAR(255) NOT NULL UNIQUE,
    account_name VARCHAR(255) NOT NULL,
    iban VARCHAR(50),
    account_number VARCHAR(50),
    currency VARCHAR(3) NOT NULL DEFAULT 'SEK',
    balance_amount DECIMAL(15, 2),
    balance_updated_at TIMESTAMP,
    mapped_account_no INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (connection_id) REFERENCES bank_connections(connection_id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (mapped_account_no) REFERENCES accounts(account_no) ON DELETE SET NULL
);

CREATE INDEX idx_bank_accounts_connection ON bank_accounts(connection_id);
CREATE INDEX idx_bank_accounts_company ON bank_accounts(company_id);
CREATE INDEX idx_bank_accounts_tink_id ON bank_accounts(tink_account_id);

-- Migration 010: Create bank_transactions table
CREATE TABLE IF NOT EXISTS bank_transactions (
    bank_transaction_id SERIAL PRIMARY KEY,
    bank_account_id INT NOT NULL,
    company_id INT NOT NULL,
    tink_transaction_id VARCHAR(255) NOT NULL UNIQUE,
    booked_date DATE NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'SEK',
    description_original TEXT NOT NULL,
    description_display TEXT,
    merchant_name VARCHAR(255),
    merchant_category_code VARCHAR(10),
    tink_category_id VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'BOOKED' CHECK (status IN ('BOOKED', 'PENDING')),
    import_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (import_status IN ('pending', 'suggested', 'approved', 'skipped', 'booked')),
    suggested_account_no INT,
    suggested_description TEXT,
    voucher_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(bank_account_id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (suggested_account_no) REFERENCES accounts(account_no) ON DELETE SET NULL,
    FOREIGN KEY (voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL
);

CREATE INDEX idx_bank_txn_company ON bank_transactions(company_id);
CREATE INDEX idx_bank_txn_account ON bank_transactions(bank_account_id);
CREATE INDEX idx_bank_txn_tink_id ON bank_transactions(tink_transaction_id);
CREATE INDEX idx_bank_txn_import_status ON bank_transactions(import_status);
CREATE INDEX idx_bank_txn_booked_date ON bank_transactions(booked_date);
CREATE INDEX idx_bank_txn_voucher ON bank_transactions(voucher_id);

-- Migration 011: Create bank_categorization_rules table
CREATE TABLE IF NOT EXISTS bank_categorization_rules (
    rule_id SERIAL PRIMARY KEY,
    company_id INT NOT NULL,
    match_pattern VARCHAR(500) NOT NULL,
    match_type VARCHAR(20) NOT NULL DEFAULT 'contains' CHECK (match_type IN ('exact', 'contains', 'regex')),
    account_no INT NOT NULL,
    description_template TEXT,
    tax_code INT,
    confidence DECIMAL(3, 2) NOT NULL DEFAULT 1.00,
    use_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE,
    FOREIGN KEY (account_no) REFERENCES accounts(account_no) ON DELETE CASCADE
);

CREATE INDEX idx_categorization_rules_company ON bank_categorization_rules(company_id);
CREATE INDEX idx_categorization_rules_pattern ON bank_categorization_rules(match_pattern);
ALTER TABLE bank_categorization_rules ADD CONSTRAINT categorization_rules_company_pattern_unique UNIQUE (company_id, match_pattern, match_type);

-- Migration 012: Create tink_oauth_state table (CSRF protection)
CREATE TABLE IF NOT EXISTS tink_oauth_state (
    state_id VARCHAR(64) PRIMARY KEY,
    company_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- Migration: Add Stripe columns to companies
ALTER TABLE companies ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(255);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS plan_status VARCHAR(20) DEFAULT 'trialing';

-- Migration 013: Create customers table
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

CREATE INDEX IF NOT EXISTS idx_customers_company ON customers(company_id);
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(company_id, name);

-- Migration 014: Create invoices table
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
    FOREIGN KEY (payment_voucher_id) REFERENCES vouchers(voucher_id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_company_number ON invoices(company_id, invoice_number);
CREATE INDEX IF NOT EXISTS idx_invoices_company ON invoices(company_id);
CREATE INDEX IF NOT EXISTS idx_invoices_customer ON invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(company_id, status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(company_id, due_date);

-- Migration 015: Create invoice_lines table
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
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE,
    FOREIGN KEY (account_no) REFERENCES accounts(account_no) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_invoice_lines_invoice ON invoice_lines(invoice_id);

-- Migration 016: Create invoice_settings table
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
