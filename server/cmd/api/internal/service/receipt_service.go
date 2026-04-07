package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"esbio/api/internal/domain"
	"fmt"
	"io"
	"net/http"
)

type ReceiptService struct {
	apiKey string
}

func NewReceiptService(apiKey string) *ReceiptService {
	return &ReceiptService{apiKey: apiKey}
}

const receiptSystemPrompt = `Du är en AI som analyserar kvitton och fakturor för ett svenskt bokföringssystem (BAS-kontoplanen).

Analysera bilden och extrahera följande information. Svara ALLTID med giltig JSON i exakt detta format:

{
  "date": "YYYY-MM-DD",
  "vendor": "Leverantörens namn",
  "description": "Kort beskrivning av köpet",
  "total_amount": 0.00,
  "vat_rate": 25,
  "vat_amount": 0.00,
  "amount_excl_vat": 0.00,
  "currency": "SEK",
  "suggested_lines": [
    {"account_no": 0, "account_name": "Kontonamn", "debit_amount": 0.00, "credit_amount": 0.00, "tax_code": 25},
    {"account_no": 0, "account_name": "Kontonamn", "debit_amount": 0.00, "credit_amount": 0.00, "tax_code": 0}
  ]
}

Regler för kontering (BAS-kontoplanen):
- Inköp/kostnader: Debet på kostnadskonto (5xxx-6xxx), Kredit på 1930 (Företagskonto/bank)
- Ingående moms 25%: Debet 2641, moms 12%: Debet 2642, moms 6%: Debet 2643
- Kontorsförbrukning: 6110
- Förbrukningsinventarier: 6150
- Datorutrustning: 5410 (IT/dator)
- Programvara/licenser: 6540
- Resor: 5800-5899
- Representation: 6071 (extern), 6072 (intern)
- Telefon/internet: 6212
- Porto/frakt: 6250
- Mat/livsmedel: 4010 (för återförsäljning) eller 6072 (intern representation)
- Lokalhyra: 6010
- Konsulttjänster: 6550
- Reklam/marknadsföring: 5910

Verifikatet MÅSTE balansera (total debet = total kredit).
Separera ALLTID momsen på eget konto (264x).
Om du inte kan avgöra momssats, anta 25%.
Om du inte kan läsa datum, använd dagens datum.
Svara BARA med JSON, ingen annan text.`

func (s *ReceiptService) ScanReceipt(imageData []byte, mimeType string) (*domain.ReceiptScanResult, error) {
	b64Image := base64.StdEncoding.EncodeToString(imageData)

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: receiptSystemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{
						InlineData: &geminiInlineData{
							MimeType: mimeType,
							Data:     b64Image,
						},
					},
					{Text: "Analysera detta kvitto/faktura och returnera JSON med extraherad data och konteringsförslag."},
				},
			},
		},
		Tools: nil,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", s.apiKey)

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Strip markdown code fences if present
	responseText = stripCodeFences(responseText)

	var result domain.ReceiptScanResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse receipt data: %w (raw: %s)", err, responseText)
	}

	return &result, nil
}

func stripCodeFences(s string) string {
	// Remove ```json ... ``` wrapping
	if len(s) > 7 && s[:7] == "```json" {
		s = s[7:]
	} else if len(s) > 3 && s[:3] == "```" {
		s = s[3:]
	}
	if len(s) > 3 && s[len(s)-3:] == "```" {
		s = s[:len(s)-3]
	}
	return s
}
