package service

import (
	"bytes"
	"cmd/api/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AgentService struct {
	apiKey               string
	voucherService       *VoucherService
	accountService       *AccountService
	lineItemService      *LineItemService
	reportService        *ReportService
	scheduledTaskService *ScheduledTaskService
}

func NewAgentService(
	apiKey string,
	voucherService *VoucherService,
	accountService *AccountService,
	lineItemService *LineItemService,
	reportService *ReportService,
	scheduledTaskService *ScheduledTaskService,
) *AgentService {
	return &AgentService{
		apiKey:               apiKey,
		voucherService:       voucherService,
		accountService:       accountService,
		lineItemService:      lineItemService,
		reportService:        reportService,
		scheduledTaskService: scheduledTaskService,
	}
}

const systemPrompt = `Du är en bokföringsassistent för Eskio, ett svenskt bokföringssystem som använder BAS-kontoplanen.

Du hjälper användare med:
- Skapa verifikat (vouchers) med korrekta konton och belopp
- Visa kontosaldon och transaktioner
- Generera rapporter (resultaträkning, balansräkning, momsrapport)
- Schemalägga återkommande verifikat

Viktiga regler:
- Alla verifikat måste balansera (debet = kredit)
- Använd svenska BAS-kontonummer (t.ex. 1930 Företagskonto, 3301 Försäljning tjänster, 2611 Utgående moms)
- Moms (VAT) i Sverige: 25%, 12%, 6%, eller 0%
- Period-format: YYYY-MM
- Datum-format: YYYY-MM-DD
- Belopp i svenska kronor (SEK)

När du skapar verifikat med moms, separera alltid momsen på eget konto (t.ex. 2611 för utgående moms 25%).

Svara alltid på svenska. Var koncis och tydlig.`

// Gemini API types
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	Tools            []geminiToolDecl       `json:"tools"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string                `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall   `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResp   `json:"functionResponse,omitempty"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiFunctionResp struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type geminiToolDecl struct {
	FunctionDeclarations []geminiFunction `json:"functionDeclarations"`
}

type geminiFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
			Role  string      `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *AgentService) getToolDeclarations() []geminiToolDecl {
	return []geminiToolDecl{
		{
			FunctionDeclarations: []geminiFunction{
				{
					Name:        "get_voucher",
					Description: "Hämta ett verifikat med dess konteringsrader baserat på verifikat-ID",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"voucher_id": map[string]interface{}{
								"type":        "integer",
								"description": "Verifikatets ID",
							},
						},
						"required": []string{"voucher_id"},
					},
				},
				{
					Name:        "get_vouchers_by_period",
					Description: "Hämta alla verifikat för en viss period (YYYY-MM)",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"period": map[string]interface{}{
								"type":        "string",
								"description": "Period i format YYYY-MM",
							},
						},
						"required": []string{"period"},
					},
				},
				{
					Name:        "get_user_vouchers",
					Description: "Hämta alla verifikat skapade av en specifik användare",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"user_id": map[string]interface{}{
								"type":        "integer",
								"description": "Användarens ID",
							},
						},
						"required": []string{"user_id"},
					},
				},
				{
					Name:        "create_voucher",
					Description: "Skapa ett nytt verifikat med konteringsrader. Debet måste vara lika med kredit.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"date": map[string]interface{}{
								"type":        "string",
								"description": "Datum (YYYY-MM-DD)",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Beskrivning av transaktionen",
							},
							"reference": map[string]interface{}{
								"type":        "string",
								"description": "Referensnummer (faktura, kvitto etc.)",
							},
							"total_amount": map[string]interface{}{
								"type":        "number",
								"description": "Totalbelopp",
							},
							"period": map[string]interface{}{
								"type":        "string",
								"description": "Period (YYYY-MM)",
							},
							"created_by": map[string]interface{}{
								"type":        "integer",
								"description": "Användar-ID som skapar verifikatet",
							},
							"line_items": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"account_no": map[string]interface{}{
											"type":        "integer",
											"description": "Kontonummer (BAS)",
										},
										"debit_amount": map[string]interface{}{
											"type":        "number",
											"description": "Debetbelopp (0 om kredit)",
										},
										"credit_amount": map[string]interface{}{
											"type":        "number",
											"description": "Kreditbelopp (0 om debet)",
										},
										"tax_code": map[string]interface{}{
											"type":        "integer",
											"description": "Momskod (25, 12, 6, eller 0)",
										},
									},
									"required": []string{"account_no", "debit_amount", "credit_amount", "tax_code"},
								},
								"description": "Konteringsrader",
							},
						},
						"required": []string{"date", "description", "reference", "total_amount", "period", "created_by", "line_items"},
					},
				},
				{
					Name:        "get_account_ledger",
					Description: "Hämta kontoutdrag (huvudbok) för ett specifikt konto med löpande saldo",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"account_no": map[string]interface{}{
								"type":        "integer",
								"description": "Kontonummer",
							},
							"period": map[string]interface{}{
								"type":        "string",
								"description": "Valfri period-filter (YYYY-MM)",
							},
						},
						"required": []string{"account_no"},
					},
				},
				{
					Name:        "search_accounts",
					Description: "Sök efter konton i kontoplanen baserat på kontonummer eller namn",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]interface{}{
								"type":        "string",
								"description": "Sökterm (kontonummer eller del av kontonamn)",
							},
						},
						"required": []string{"query"},
					},
				},
				{
					Name:        "get_income_statement",
					Description: "Hämta resultaträkning för en period",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"from_date": map[string]interface{}{
								"type":        "string",
								"description": "Startdatum (YYYY-MM-DD)",
							},
							"to_date": map[string]interface{}{
								"type":        "string",
								"description": "Slutdatum (YYYY-MM-DD)",
							},
						},
						"required": []string{"from_date", "to_date"},
					},
				},
				{
					Name:        "get_balance_sheet",
					Description: "Hämta balansräkning per ett visst datum",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"as_of_date": map[string]interface{}{
								"type":        "string",
								"description": "Datum (YYYY-MM-DD)",
							},
						},
						"required": []string{"as_of_date"},
					},
				},
				{
					Name:        "create_scheduled_task",
					Description: "Schemalägga ett återkommande verifikat som skapas automatiskt varje månad på en viss dag",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"user_id": map[string]interface{}{
								"type":        "integer",
								"description": "Användar-ID",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Beskrivning av den schemalagda uppgiften",
							},
							"prompt": map[string]interface{}{
								"type":        "string",
								"description": "Instruktionen som agenten ska utföra varje gång",
							},
							"template_voucher_id": map[string]interface{}{
								"type":        "integer",
								"description": "Valfritt: ID för ett befintligt verifikat att använda som mall",
							},
							"day_of_month": map[string]interface{}{
								"type":        "integer",
								"description": "Dag i månaden (1-31) då uppgiften ska köras",
							},
						},
						"required": []string{"user_id", "description", "prompt", "day_of_month"},
					},
				},
				{
					Name:        "list_scheduled_tasks",
					Description: "Lista alla schemalagda uppgifter för en användare",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"user_id": map[string]interface{}{
								"type":        "integer",
								"description": "Användar-ID",
							},
						},
						"required": []string{"user_id"},
					},
				},
			},
		},
	}
}

func (s *AgentService) executeTool(name string, args map[string]interface{}) (string, error) {
	// Convert args to JSON for consistent parsing
	input, _ := json.Marshal(args)

	switch name {
	case "get_voucher":
		var params struct {
			VoucherID int `json:"voucher_id"`
		}
		json.Unmarshal(input, &params)
		// Handle float -> int from Gemini
		if params.VoucherID == 0 {
			if v, ok := args["voucher_id"].(float64); ok {
				params.VoucherID = int(v)
			}
		}
		voucher, err := s.voucherService.GetVoucherByID(params.VoucherID)
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		data, _ := json.Marshal(voucher)
		return string(data), nil

	case "get_vouchers_by_period":
		period, _ := args["period"].(string)
		vouchers, err := s.voucherService.GetVouchersByPeriod(period)
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		data, _ := json.Marshal(vouchers)
		return string(data), nil

	case "get_user_vouchers":
		userID := intFromArg(args["user_id"])
		vouchers, err := s.voucherService.GetVouchersByCreatedBy(userID)
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		data, _ := json.Marshal(vouchers)
		return string(data), nil

	case "create_voucher":
		dateStr, _ := args["date"].(string)
		description, _ := args["description"].(string)
		reference, _ := args["reference"].(string)
		totalAmount := floatFromArg(args["total_amount"])
		period, _ := args["period"].(string)
		createdBy := intFromArg(args["created_by"])

		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Sprintf("Fel: ogiltigt datum %s", dateStr), nil
		}

		voucher := &domain.Voucher{
			Date:        domain.FlexibleDate{Time: t},
			Description: description,
			Reference:   reference,
			TotalAmount: totalAmount,
			Period:      period,
			CreatedBy:   createdBy,
		}

		if err := s.voucherService.CreateVoucher(voucher); err != nil {
			return fmt.Sprintf("Fel vid skapande av verifikat: %s", err.Error()), nil
		}

		// Create line items
		lineItemsRaw, ok := args["line_items"].([]interface{})
		if ok {
			for _, liRaw := range lineItemsRaw {
				li, ok := liRaw.(map[string]interface{})
				if !ok {
					continue
				}
				lineItem := &domain.LineItem{
					VoucherID:    voucher.VoucherID,
					AccountNo:    intFromArg(li["account_no"]),
					DebitAmount:  floatFromArg(li["debit_amount"]),
					CreditAmount: floatFromArg(li["credit_amount"]),
					TaxCode:      intFromArg(li["tax_code"]),
				}
				if err := s.lineItemService.CreateLineItem(lineItem); err != nil {
					return fmt.Sprintf("Verifikat #%d skapat men fel vid konteringsrad: %s", voucher.VoucherNumber, err.Error()), nil
				}
			}
		}

		return fmt.Sprintf("Verifikat #%d skapat (ID: %d) med %d konteringsrader. Datum: %s, Belopp: %.2f kr",
			voucher.VoucherNumber, voucher.VoucherID, len(lineItemsRaw), dateStr, totalAmount), nil

	case "get_account_ledger":
		accountNo := intFromArg(args["account_no"])
		period, _ := args["period"].(string)
		ledger, err := s.accountService.GetLedger(accountNo, period)
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		data, _ := json.Marshal(ledger)
		return string(data), nil

	case "search_accounts":
		query, _ := args["query"].(string)
		accounts, err := s.accountService.GetAllAccounts()
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		var results []domain.Account
		q := strings.ToLower(query)
		for _, acc := range accounts {
			if strings.Contains(strings.ToLower(acc.AccountName), q) ||
				strings.Contains(strconv.Itoa(acc.AccountNo), q) {
				results = append(results, *acc)
			}
		}
		if len(results) > 20 {
			results = results[:20]
		}
		data, _ := json.Marshal(results)
		return string(data), nil

	case "get_income_statement":
		fromDate, _ := args["from_date"].(string)
		toDate, _ := args["to_date"].(string)
		statement, err := s.reportService.GetIncomeStatement(fromDate, toDate)
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		data, _ := json.Marshal(statement)
		return string(data), nil

	case "get_balance_sheet":
		asOfDate, _ := args["as_of_date"].(string)
		sheet, err := s.reportService.GetBalanceSheet(asOfDate)
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		data, _ := json.Marshal(sheet)
		return string(data), nil

	case "create_scheduled_task":
		userID := intFromArg(args["user_id"])
		description, _ := args["description"].(string)
		prompt, _ := args["prompt"].(string)
		dayOfMonth := intFromArg(args["day_of_month"])

		task := &domain.ScheduledTask{
			UserID:      userID,
			Description: description,
			Prompt:      prompt,
			DayOfMonth:  dayOfMonth,
			Active:      true,
		}

		if v, ok := args["template_voucher_id"]; ok {
			tid := intFromArg(v)
			if tid > 0 {
				task.TemplateVoucherID = &tid
			}
		}

		if err := s.scheduledTaskService.CreateTask(task); err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		return fmt.Sprintf("Schemalagd uppgift skapad (ID: %d). '%s' körs den %d:e varje månad.",
			task.TaskID, description, dayOfMonth), nil

	case "list_scheduled_tasks":
		userID := intFromArg(args["user_id"])
		tasks, err := s.scheduledTaskService.GetTasksByUserID(userID)
		if err != nil {
			return fmt.Sprintf("Fel: %s", err.Error()), nil
		}
		data, _ := json.Marshal(tasks)
		return string(data), nil

	default:
		return fmt.Sprintf("Okänt verktyg: %s", name), nil
	}
}

func (s *AgentService) Chat(userID int, conversationID string, userMessage string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not configured")
	}

	contents := []geminiContent{
		{
			Role:  "user",
			Parts: []geminiPart{{Text: userMessage}},
		},
	}

	// Tool use loop
	for i := 0; i < 10; i++ {
		reqBody := geminiRequest{
			Contents: contents,
			Tools:    s.getToolDeclarations(),
			SystemInstruction: &geminiContent{
				Parts: []geminiPart{{Text: systemPrompt}},
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request: %w", err)
		}

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", s.apiKey)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 120 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("API request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return "", fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(body))
		}

		var geminiResp geminiResponse
		if err := json.Unmarshal(body, &geminiResp); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}

		if geminiResp.Error != nil {
			return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
		}

		if len(geminiResp.Candidates) == 0 {
			return "Inget svar från agenten.", nil
		}

		candidate := geminiResp.Candidates[0]
		parts := candidate.Content.Parts

		// Check for function calls
		var hasFunctionCall bool
		var textResult strings.Builder

		for _, part := range parts {
			if part.FunctionCall != nil {
				hasFunctionCall = true
			}
			if part.Text != "" {
				textResult.WriteString(part.Text)
			}
		}

		if !hasFunctionCall {
			return textResult.String(), nil
		}

		// Add model response to contents
		contents = append(contents, geminiContent{
			Role:  "model",
			Parts: parts,
		})

		// Execute function calls and add results
		var responseParts []geminiPart
		for _, part := range parts {
			if part.FunctionCall != nil {
				result, err := s.executeTool(part.FunctionCall.Name, part.FunctionCall.Args)
				if err != nil {
					result = fmt.Sprintf("Fel: %s", err.Error())
				}

				responseParts = append(responseParts, geminiPart{
					FunctionResponse: &geminiFunctionResp{
						Name: part.FunctionCall.Name,
						Response: map[string]interface{}{
							"result": result,
						},
					},
				})
			}
		}

		contents = append(contents, geminiContent{
			Role:  "user",
			Parts: responseParts,
		})
	}

	return "Jag kunde inte slutföra uppgiften efter flera försök.", nil
}

// Helper functions for Gemini's JSON number handling
func intFromArg(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case json.Number:
		i, _ := val.Int64()
		return int(i)
	default:
		return 0
	}
}

func floatFromArg(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}
