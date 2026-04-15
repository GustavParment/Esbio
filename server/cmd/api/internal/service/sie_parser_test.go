package service

import (
	"strings"
	"testing"
)

const sampleSIE = `#FLAGGA 0
#FORMAT PC8
#SIETYP 4
#PROGRAM "Fortnox" "1.0"
#GEN 20250115
#FNAMN "Test AB"
#ORGNR 556677-8899
#KPTYP BAS2024
#RAR 0 20250101 20251231

#KONTO 1930 "Företagskonto"
#KONTO 3010 "Försäljning 25%"
#KONTO 2611 "Utgående moms 25%"

#IB 0 1930 0.00

#VER "A" 1 20250115 "Faktura 100"
{
  #TRANS 1930 {} 1250.00
  #TRANS 3010 {} -1000.00
  #TRANS 2611 {} -250.00
}

#VER "A" 2 20250120 "Inköp"
{
  #TRANS 1930 {} -500.00
  #TRANS 3010 {} 500.00
}
`

func TestParseSIE_HappyPath(t *testing.T) {
	r, err := ParseSIE([]byte(sampleSIE))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.CompanyName != "Test AB" {
		t.Errorf("CompanyName = %q, want %q", r.CompanyName, "Test AB")
	}
	if r.OrgNumber != "556677-8899" {
		t.Errorf("OrgNumber = %q", r.OrgNumber)
	}
	if r.FiscalFrom != "2025-01-01" || r.FiscalTo != "2025-12-31" {
		t.Errorf("Fiscal = %s..%s", r.FiscalFrom, r.FiscalTo)
	}
	if len(r.Accounts) != 3 {
		t.Errorf("Accounts = %d, want 3", len(r.Accounts))
	}
	if len(r.Vouchers) != 2 {
		t.Fatalf("Vouchers = %d, want 2", len(r.Vouchers))
	}

	v1 := r.Vouchers[0]
	if v1.Description != "Faktura 100" {
		t.Errorf("Voucher 1 description = %q", v1.Description)
	}
	if len(v1.Lines) != 3 {
		t.Errorf("Voucher 1 lines = %d, want 3", len(v1.Lines))
	}

	// Voucher 1 should be balanced (no warning), voucher 2 is balanced too
	for _, w := range r.Warnings {
		if strings.Contains(w, "unbalanced") {
			t.Errorf("unexpected unbalanced warning: %s", w)
		}
	}
}

func TestParseSIE_DetectsUnbalanced(t *testing.T) {
	bad := `#FNAMN "X"
#KONTO 1930 "Bank"
#KONTO 3010 "Försäljning"
#VER "A" 1 20250101 "Bad"
{
  #TRANS 1930 {} 100.00
  #TRANS 3010 {} -50.00
}
`
	r, err := ParseSIE([]byte(bad))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	found := false
	for _, w := range r.Warnings {
		if strings.Contains(w, "unbalanced") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unbalanced warning, got %v", r.Warnings)
	}
}

func TestTokenizeSIELine_QuotedStrings(t *testing.T) {
	tokens, err := tokenizeSIELine(`#VER "A" 1 20250101 "Faktura nr \"100\" till kund"`)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(tokens) != 5 {
		t.Fatalf("tokens = %d (%v)", len(tokens), tokens)
	}
	if tokens[4] != `Faktura nr "100" till kund` {
		t.Errorf("desc token = %q", tokens[4])
	}
}

func TestTokenizeSIELine_BraceBlock(t *testing.T) {
	tokens, err := tokenizeSIELine(`#TRANS 1930 {} 1000.00`)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(tokens) != 4 || tokens[2] != "{}" {
		t.Errorf("tokens = %v", tokens)
	}
}

func TestParseSIE_RejectsNonSIE(t *testing.T) {
	_, err := ParseSIE([]byte("hello world\nthis is not SIE\n"))
	if err == nil {
		t.Error("expected error for non-SIE input")
	}
}

func TestBuildAccountFromBAS(t *testing.T) {
	cases := []struct {
		no       int
		wantType string
		wantSide string
		wantGrp  int
	}{
		{1930, "BS", "Debit", 1},
		{2611, "BS", "Credit", 2},
		{3010, "P&L", "Credit", 3},
		{4010, "P&L", "Debit", 4},
		{6212, "P&L", "Debit", 6},
		{8310, "P&L", "Debit", 8},
	}
	for _, tc := range cases {
		a := buildAccountFromBAS(tc.no, "x")
		if a.Type != tc.wantType || a.StandardSide != tc.wantSide || a.AccountGroup != tc.wantGrp {
			t.Errorf("acc %d: got %s/%s/grp%d, want %s/%s/grp%d",
				tc.no, a.Type, a.StandardSide, a.AccountGroup, tc.wantType, tc.wantSide, tc.wantGrp)
		}
	}
}
