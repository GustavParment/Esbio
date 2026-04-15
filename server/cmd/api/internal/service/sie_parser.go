package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
)

// ParsedSIE is the in-memory result of parsing a SIE 4 file.
// All amounts use the SIE convention: positive = debit, negative = credit.
type ParsedSIE struct {
	CompanyName string
	OrgNumber   string
	ChartType   string         // e.g. "BAS2024"
	FiscalFrom  string         // YYYY-MM-DD
	FiscalTo    string         // YYYY-MM-DD
	Encoding    string         // detected encoding label
	Accounts    []ParsedAccount
	Vouchers    []ParsedVoucher
	Warnings    []string
}

type ParsedAccount struct {
	AccountNo   int
	AccountName string
}

type ParsedVoucher struct {
	Series      string // e.g. "A"
	Number      int    // SIE source voucher number (informational only)
	Date        time.Time
	Description string
	Lines       []ParsedTrans
}

type ParsedTrans struct {
	AccountNo int
	Amount    decimal.Decimal // SIE convention: + debit, - credit
}

// ParseSIE decodes the bytes (auto-detecting encoding) and parses a SIE 4 file.
func ParseSIE(raw []byte) (*ParsedSIE, error) {
	text, encoding := decodeSIE(raw)

	result := &ParsedSIE{Encoding: encoding}

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var currentVoucher *ParsedVoucher
	inVerBlock := false
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Block delimiters for #VER blocks
		if line == "{" {
			continue
		}
		if line == "}" {
			if currentVoucher != nil {
				result.Vouchers = append(result.Vouchers, *currentVoucher)
				currentVoucher = nil
			}
			inVerBlock = false
			continue
		}

		tokens, err := tokenizeSIELine(line)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: %v", lineNo, err))
			continue
		}
		if len(tokens) == 0 {
			continue
		}
		label := strings.ToUpper(tokens[0])

		switch label {
		case "#FNAMN":
			if len(tokens) >= 2 {
				result.CompanyName = tokens[1]
			}
		case "#ORGNR":
			if len(tokens) >= 2 {
				result.OrgNumber = tokens[1]
			}
		case "#KPTYP":
			if len(tokens) >= 2 {
				result.ChartType = tokens[1]
			}
		case "#RAR":
			// #RAR yearOffset fromYYYYMMDD toYYYYMMDD — only capture year 0 (current)
			if len(tokens) >= 4 && tokens[1] == "0" {
				if from, err := parseSIEDate(tokens[2]); err == nil {
					result.FiscalFrom = from.Format("2006-01-02")
				}
				if to, err := parseSIEDate(tokens[3]); err == nil {
					result.FiscalTo = to.Format("2006-01-02")
				}
			}
		case "#KONTO":
			// #KONTO accountNo "name"
			if len(tokens) >= 3 {
				accNo, err := strconv.Atoi(tokens[1])
				if err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: invalid account number %q", lineNo, tokens[1]))
					continue
				}
				result.Accounts = append(result.Accounts, ParsedAccount{
					AccountNo:   accNo,
					AccountName: tokens[2],
				})
			}
		case "#VER":
			// #VER series number date "description" [regdate] [signature]
			if len(tokens) < 4 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: malformed #VER", lineNo))
				continue
			}
			series := tokens[1]
			number, _ := strconv.Atoi(tokens[2])
			date, err := parseSIEDate(tokens[3])
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: invalid #VER date %q", lineNo, tokens[3]))
				continue
			}
			description := ""
			if len(tokens) >= 5 {
				description = tokens[4]
			}
			currentVoucher = &ParsedVoucher{
				Series:      series,
				Number:      number,
				Date:        date,
				Description: description,
			}
			inVerBlock = true
		case "#TRANS":
			// #TRANS accountNo {dim} amount [date] [text]
			if !inVerBlock || currentVoucher == nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: #TRANS outside #VER block", lineNo))
				continue
			}
			if len(tokens) < 4 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: malformed #TRANS", lineNo))
				continue
			}
			accNo, err := strconv.Atoi(tokens[1])
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: invalid #TRANS account %q", lineNo, tokens[1]))
				continue
			}
			// tokens[2] is the dimension object string (often "{}")
			amount, err := parseSIEAmount(tokens[3])
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: invalid #TRANS amount %q", lineNo, tokens[3]))
				continue
			}
			currentVoucher.Lines = append(currentVoucher.Lines, ParsedTrans{
				AccountNo: accNo,
				Amount:    amount,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Flush dangling voucher (file ended without closing brace — be lenient)
	if currentVoucher != nil {
		result.Warnings = append(result.Warnings, "file ended inside #VER block")
		result.Vouchers = append(result.Vouchers, *currentVoucher)
	}

	// Validate balance per voucher
	for i, v := range result.Vouchers {
		sum := decimal.Zero
		for _, l := range v.Lines {
			sum = sum.Add(l.Amount)
		}
		if !sum.IsZero() {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("voucher #%d (%s) is unbalanced (sum=%s)", i+1, v.Description, sum.StringFixed(2)))
		}
	}

	if result.CompanyName == "" && len(result.Vouchers) == 0 && len(result.Accounts) == 0 {
		return nil, errors.New("file does not look like a SIE 4 file (no header, accounts or vouchers found)")
	}

	return result, nil
}

// decodeSIE returns the file as a UTF-8 string, picking the encoding based on
// the #FORMAT directive when present, falling back to ISO-8859-1.
// SIE 4 uses PC8 (CP437) by default; some exporters write ISO-8859-1.
func decodeSIE(raw []byte) (string, string) {
	// If the file is valid UTF-8, prefer that — many modern exporters write UTF-8
	// even when the #FORMAT directive still says PC8.
	if utf8.Valid(raw) && bytes.ContainsAny(raw, "åäöÅÄÖ") {
		return string(raw), "UTF-8"
	}

	// Look for #FORMAT directive in the first 1KB
	head := raw
	if len(head) > 1024 {
		head = head[:1024]
	}
	headStr := string(head) // may be garbled but the directive itself is ASCII

	if strings.Contains(strings.ToUpper(headStr), "#FORMAT PC8") {
		decoded, err := charmap.CodePage437.NewDecoder().Bytes(raw)
		if err == nil {
			return string(decoded), "PC8 (CP437)"
		}
	}

	// Default: ISO-8859-1 (Latin-1) — every byte is a valid code point
	decoded, err := charmap.ISO8859_1.NewDecoder().Bytes(raw)
	if err == nil {
		return string(decoded), "ISO-8859-1"
	}

	return string(raw), "raw"
}

// tokenizeSIELine splits a SIE line into tokens. Quoted strings are kept as
// a single token without their surrounding quotes; backslash escapes \" and \\.
// Curly-brace dimension objects (e.g. "{}" or "{1 \"foo\"}") are returned as
// a single token including the braces.
func tokenizeSIELine(line string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inQuote := false
	inBrace := 0
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' && inQuote {
			escaped = true
			continue
		}
		if ch == '"' && inBrace == 0 {
			if inQuote {
				// closing quote — flush even if empty so we keep ""
				tokens = append(tokens, current.String())
				current.Reset()
				inQuote = false
			} else {
				flush()
				inQuote = true
			}
			continue
		}
		if inQuote {
			current.WriteByte(ch)
			continue
		}
		if ch == '{' {
			if inBrace == 0 {
				flush()
			}
			inBrace++
			current.WriteByte(ch)
			continue
		}
		if ch == '}' {
			current.WriteByte(ch)
			inBrace--
			if inBrace == 0 {
				flush()
			}
			continue
		}
		if inBrace > 0 {
			current.WriteByte(ch)
			continue
		}
		if ch == ' ' || ch == '\t' {
			flush()
			continue
		}
		current.WriteByte(ch)
	}
	if inQuote {
		return nil, errors.New("unterminated quoted string")
	}
	if inBrace > 0 {
		return nil, errors.New("unterminated brace block")
	}
	flush()
	return tokens, nil
}

func parseSIEDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if len(s) != 8 {
		return time.Time{}, fmt.Errorf("expected YYYYMMDD, got %q", s)
	}
	return time.Parse("20060102", s)
}

func parseSIEAmount(s string) (decimal.Decimal, error) {
	// SIE typically uses '.' as decimal separator; accept ',' too.
	return decimal.NewFromString(strings.ReplaceAll(s, ",", "."))
}
