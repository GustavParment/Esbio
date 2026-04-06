-- Reset sequence
ALTER SEQUENCE vouchers_number_seq RESTART WITH 1;

-- December 2025
INSERT INTO vouchers (voucher_number, date, description, reference, total_amount, period, created_by) VALUES
(nextval('vouchers_number_seq'), '2025-12-01', 'Lokalhyra december', 'Faktura 2025-1201', 15000, '2025-12', 2),
(nextval('vouchers_number_seq'), '2025-12-05', 'Inköp kontorsmaterial', 'Kvitto 2025-1205', 3500, '2025-12', 2),
(nextval('vouchers_number_seq'), '2025-12-15', 'Konsultarvode webbdesign', 'Faktura K-2025-42', 45000, '2025-12', 2),
(nextval('vouchers_number_seq'), '2025-12-20', 'Försäljning tjänster december', 'Faktura F-2025-88', 85000, '2025-12', 2),
(nextval('vouchers_number_seq'), '2025-12-25', 'Utbetalning löner december', 'Lönespec 2025-12', 52000, '2025-12', 2),
(nextval('vouchers_number_seq'), '2025-12-28', 'Arbetsgivaravgifter december', 'AGI 2025-12', 16328, '2025-12', 2),
(nextval('vouchers_number_seq'), '2025-12-30', 'Telefonkostnader december', 'Faktura T-2025-12', 2500, '2025-12', 2);

-- January 2026
INSERT INTO vouchers (voucher_number, date, description, reference, total_amount, period, created_by) VALUES
(nextval('vouchers_number_seq'), '2026-01-02', 'Lokalhyra januari', 'Faktura 2026-0101', 15000, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-08', 'Inköp kontorsmaterial', 'Kvitto 2026-0108', 4200, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-10', 'Försäljning tjänster januari', 'Faktura F-2026-01', 92000, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-15', 'Konsultarvode IT-system', 'Faktura K-2026-05', 38000, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-20', 'Försäljning varor', 'Faktura F-2026-02', 28000, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-25', 'Utbetalning löner januari', 'Lönespec 2026-01', 52000, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-25', 'Arbetsgivaravgifter januari', 'AGI 2026-01', 16328, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-28', 'Företagsförsäkring Q1', 'Faktura F-2026-Q1', 8500, '2026-01', 2),
(nextval('vouchers_number_seq'), '2026-01-30', 'Telefonkostnader januari', 'Faktura T-2026-01', 2500, '2026-01', 2);

-- February 2026
INSERT INTO vouchers (voucher_number, date, description, reference, total_amount, period, created_by) VALUES
(nextval('vouchers_number_seq'), '2026-02-01', 'Lokalhyra februari', 'Faktura 2026-0201', 15000, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-05', 'Inköp kontorsmaterial och skrivare', 'Kvitto 2026-0205', 8900, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-10', 'Försäljning tjänster februari', 'Faktura F-2026-10', 75000, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-12', 'Konsultarvode redovisning', 'Faktura K-2026-12', 22000, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-14', 'Resekostnader kundbesök', 'Kvitto R-2026-02', 5600, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-15', 'Försäljning varor februari', 'Faktura F-2026-15', 35000, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-18', 'Utbetalning löner februari', 'Lönespec 2026-02', 52000, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-18', 'Arbetsgivaravgifter februari', 'AGI 2026-02', 16328, '2026-02', 2),
(nextval('vouchers_number_seq'), '2026-02-19', 'Bankkostnader februari', 'Bank 2026-02', 350, '2026-02', 2);

-- Line items (no description column)
INSERT INTO line_items (voucher_id, account_no, debit_amount, credit_amount) VALUES
-- V1: Lokalhyra dec
(1, 6010, 15000, 0), (1, 1930, 0, 15000),
-- V2: Kontorsmaterial dec
(2, 6110, 3500, 0), (2, 1930, 0, 3500),
-- V3: Konsultarvode dec
(3, 6550, 45000, 0), (3, 2440, 0, 45000),
-- V4: Försäljning tjänster dec
(4, 1510, 85000, 0), (4, 3040, 0, 68000), (4, 2610, 0, 17000),
-- V5: Löner dec
(5, 5010, 52000, 0), (5, 1930, 0, 36400), (5, 2710, 0, 15600),
-- V6: AGI dec
(6, 5210, 16328, 0), (6, 2730, 0, 16328),
-- V7: Telefon dec
(7, 6200, 2500, 0), (7, 1930, 0, 2500),
-- V8: Lokalhyra jan
(8, 6010, 15000, 0), (8, 1930, 0, 15000),
-- V9: Kontorsmaterial jan
(9, 6110, 4200, 0), (9, 1930, 0, 4200),
-- V10: Försäljning tjänster jan
(10, 1510, 92000, 0), (10, 3040, 0, 73600), (10, 2610, 0, 18400),
-- V11: Konsultarvode jan
(11, 6550, 38000, 0), (11, 2440, 0, 38000),
-- V12: Försäljning varor jan
(12, 1510, 28000, 0), (12, 3010, 0, 22400), (12, 2610, 0, 5600),
-- V13: Löner jan
(13, 5010, 52000, 0), (13, 1930, 0, 36400), (13, 2710, 0, 15600),
-- V14: AGI jan
(14, 5210, 16328, 0), (14, 2730, 0, 16328),
-- V15: Försäkring Q1
(15, 6300, 8500, 0), (15, 1930, 0, 8500),
-- V16: Telefon jan
(16, 6200, 2500, 0), (16, 1930, 0, 2500),
-- V17: Lokalhyra feb
(17, 6010, 15000, 0), (17, 1930, 0, 15000),
-- V18: Kontorsmaterial + skrivare feb
(18, 6110, 8900, 0), (18, 1930, 0, 8900),
-- V19: Försäljning tjänster feb
(19, 1510, 75000, 0), (19, 3040, 0, 60000), (19, 2610, 0, 15000),
-- V20: Konsultarvode redovisning feb
(20, 6550, 22000, 0), (20, 2440, 0, 22000),
-- V21: Resekostnader feb
(21, 6700, 5600, 0), (21, 1930, 0, 5600),
-- V22: Försäljning varor feb
(22, 1510, 35000, 0), (22, 3010, 0, 28000), (22, 2610, 0, 7000),
-- V23: Löner feb
(23, 5010, 52000, 0), (23, 1930, 0, 36400), (23, 2710, 0, 15600),
-- V24: AGI feb
(24, 5210, 16328, 0), (24, 2730, 0, 16328),
-- V25: Bankkostnader feb
(25, 6570, 350, 0), (25, 1930, 0, 350);
