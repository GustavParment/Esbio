package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type EmailService struct {
	apiKey  string
	fromAddr string
}

func NewEmailService(apiKey string) *EmailService {
	return &EmailService{
		apiKey:   apiKey,
		fromAddr: "noreply@esbio.se",
	}
}

type emailAttachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type emailRequest struct {
	From        string            `json:"from"`
	To          []string          `json:"to"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Subject     string            `json:"subject"`
	HTML        string            `json:"html"`
	Attachments []emailAttachment `json:"attachments,omitempty"`
}

type emailResponse struct {
	ID string `json:"id"`
}

func (s *EmailService) SendInvoiceEmail(toEmail, customerName, companyName string, invoiceNumber int, total string, dueDate string, pdfBytes []byte) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("RESEND_API_KEY is not configured")
	}

	subject := fmt.Sprintf("Faktura %d från %s", invoiceNumber, companyName)

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Faktura %d</h2>
			<p>Hej %s,</p>
			<p>Bifogat finner du faktura <strong>%d</strong> från %s.</p>
			<table style="margin: 20px 0; border-collapse: collapse;">
				<tr>
					<td style="padding: 4px 16px 4px 0; color: #666;">Belopp:</td>
					<td style="padding: 4px 0;"><strong>%s kr</strong></td>
				</tr>
				<tr>
					<td style="padding: 4px 16px 4px 0; color: #666;">Förfallodatum:</td>
					<td style="padding: 4px 0;">%s</td>
				</tr>
			</table>
			<p>Se bifogad PDF för fullständig faktura och betalningsinformation.</p>
			<p style="margin-top: 20px;">Har du frågor om denna faktura? Kontakta oss på <a href="mailto:info@esbio.se" style="color: #16a34a;">info@esbio.se</a></p>
			<hr style="border: none; border-top: 1px solid #eee; margin: 24px 0;">
			<p style="color: #999; font-size: 12px;">Skickat via Esbio Bokföringssystem</p>
		</div>
	`, invoiceNumber, customerName, invoiceNumber, companyName, total, dueDate)

	req := emailRequest{
		From:    fmt.Sprintf("%s <%s>", companyName, s.fromAddr),
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
	}

	if len(pdfBytes) > 0 {
		req.Attachments = []emailAttachment{
			{
				Filename: fmt.Sprintf("faktura_%d.pdf", invoiceNumber),
				Content:  base64.StdEncoding.EncodeToString(pdfBytes),
			},
		}
	}

	return s.send(req)
}

func (s *EmailService) SendVerificationEmail(toEmail, firstName, verifyURL string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("RESEND_API_KEY is not configured")
	}

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Välkommen till Esbio!</h2>
			<p>Hej %s,</p>
			<p>Tack för att du registrerade dig. Vänligen verifiera din e-postadress genom att klicka på knappen nedan:</p>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #16a34a; color: white; padding: 12px 32px; text-decoration: none; border-radius: 8px; font-weight: bold; display: inline-block;">Verifiera e-post</a>
			</div>
			<p style="color: #666; font-size: 14px;">Länken är giltig i 24 timmar. Om du inte registrerade dig kan du ignorera detta mail.</p>
			<hr style="border: none; border-top: 1px solid #eee; margin: 24px 0;">
			<p style="color: #999; font-size: 12px;">Parment Software Solutions AB</p>
		</div>
	`, firstName, verifyURL)

	req := emailRequest{
		From:    fmt.Sprintf("Esbio <%s>", s.fromAddr),
		To:      []string{toEmail},
		Subject: "Verifiera din e-postadress — Esbio",
		HTML:    html,
	}

	return s.send(req)
}

func (s *EmailService) SendWelcomeEmail(toEmail, firstName string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("RESEND_API_KEY is not configured")
	}

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Välkommen till Esbio, %s!</h2>
			<p>Din e-postadress är nu verifierad och ditt konto är redo att använda.</p>
			<p>Med Esbio kan du:</p>
			<ul style="color: #333; line-height: 1.8;">
				<li>Bokföra verifikat enkelt och snabbt</li>
				<li>Skapa och skicka fakturor</li>
				<li>Få översikt över din ekonomi med rapporter</li>
			</ul>
			<div style="text-align: center; margin: 30px 0;">
				<a href="https://app.esbio.se" style="background-color: #16a34a; color: white; padding: 12px 32px; text-decoration: none; border-radius: 8px; font-weight: bold; display: inline-block;">Logga in</a>
			</div>
			<p>Har du frågor? Kontakta oss på <a href="mailto:info@esbio.se" style="color: #16a34a;">info@esbio.se</a></p>
			<hr style="border: none; border-top: 1px solid #eee; margin: 24px 0;">
			<p style="color: #999; font-size: 12px;">Parment Software Solutions AB</p>
		</div>
	`, firstName)

	req := emailRequest{
		From:    fmt.Sprintf("Esbio <%s>", s.fromAddr),
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Välkommen till Esbio, %s!", firstName),
		HTML:    html,
	}

	return s.send(req)
}

func (s *EmailService) SendPasswordResetEmail(toEmail, firstName, resetURL string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("RESEND_API_KEY is not configured")
	}

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Återställ lösenord</h2>
			<p>Hej %s,</p>
			<p>Vi har fått en begäran om att återställa lösenordet för ditt Esbio-konto. Klicka på knappen nedan för att välja ett nytt lösenord:</p>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #16a34a; color: white; padding: 12px 32px; text-decoration: none; border-radius: 8px; font-weight: bold; display: inline-block;">Återställ lösenord</a>
			</div>
			<p style="color: #666; font-size: 14px;">Länken är giltig i 1 timme. Om du inte begärde detta kan du ignorera detta mail.</p>
			<hr style="border: none; border-top: 1px solid #eee; margin: 24px 0;">
			<p style="color: #999; font-size: 12px;">Parment Software Solutions AB</p>
		</div>
	`, firstName, resetURL)

	req := emailRequest{
		From:    fmt.Sprintf("Esbio <%s>", s.fromAddr),
		To:      []string{toEmail},
		Subject: "Återställ ditt lösenord — Esbio",
		HTML:    html,
	}

	return s.send(req)
}

func (s *EmailService) SendSupportEmail(fromName, fromEmail, companyName, subject, message string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("RESEND_API_KEY is not configured")
	}

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Supportärende</h2>
			<table style="border-collapse: collapse; margin: 16px 0;">
				<tr>
					<td style="padding: 4px 16px 4px 0; color: #666; vertical-align: top;">Från:</td>
					<td style="padding: 4px 0;">%s &lt;%s&gt;</td>
				</tr>
				<tr>
					<td style="padding: 4px 16px 4px 0; color: #666; vertical-align: top;">Företag:</td>
					<td style="padding: 4px 0;">%s</td>
				</tr>
			</table>
			<hr style="border: none; border-top: 1px solid #eee; margin: 16px 0;">
			<div style="white-space: pre-wrap; line-height: 1.6;">%s</div>
		</div>
	`, fromName, fromEmail, companyName, message)

	req := emailRequest{
		From:    fmt.Sprintf("Esbio Support <%s>", s.fromAddr),
		To:      []string{"info@esbio.se"},
		Subject: fmt.Sprintf("[Support] %s", subject),
		HTML:    html,
		ReplyTo: fromEmail,
	}

	return s.send(req)
}

func (s *EmailService) send(req emailRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal email request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resend API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var emailResp emailResponse
	if err := json.Unmarshal(respBody, &emailResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return emailResp.ID, nil
}
