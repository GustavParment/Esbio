package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceEmailHandler struct {
	invoiceService *service.InvoiceService
	companyService *service.CompanyService
	emailService   *service.EmailService
}

func NewInvoiceEmailHandler(invoiceService *service.InvoiceService, companyService *service.CompanyService, emailService *service.EmailService) *InvoiceEmailHandler {
	return &InvoiceEmailHandler{
		invoiceService: invoiceService,
		companyService: companyService,
		emailService:   emailService,
	}
}

// SendInvoiceEmail handles POST /invoices/:id/send-email
func (h *InvoiceEmailHandler) SendInvoiceEmail(c *gin.Context) {
	invoiceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
		return
	}

	companyID, _ := c.Get("companyID")

	invoice, err := h.invoiceService.GetInvoiceWithDetails(invoiceID)
	if err != nil {
		notFoundError(c, err)
		return
	}
	if invoice.CompanyID != companyID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if invoice.Status == "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kan inte skicka utkast — fakturan måste finaliseras först"})
		return
	}

	if invoice.Customer == nil || invoice.Customer.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kunden saknar e-postadress"})
		return
	}

	company, err := h.companyService.GetCompanyByID(invoice.CompanyID)
	if err != nil {
		internalError(c, err)
		return
	}

	pdfBytes, err := buildInvoicePDF(invoice, company)
	if err != nil {
		internalError(c, err)
		return
	}

	emailID, err := h.emailService.SendInvoiceEmail(
		invoice.Customer.Email,
		invoice.Customer.Name,
		company.CompanyName,
		invoice.InvoiceNumber,
		invoice.Total.StringFixed(2),
		invoice.DueDate.Time.Format("2006-01-02"),
		pdfBytes,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to send invoice email for invoice %d: %v", invoiceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kunde inte skicka e-post"})
		return
	}

	log.Printf("[INFO] Invoice %d emailed to %s (resend ID: %s)", invoice.InvoiceNumber, invoice.Customer.Email, emailID)

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Faktura %d skickad till %s", invoice.InvoiceNumber, invoice.Customer.Email),
		"email_id": emailID,
	})
}

// buildInvoicePDF generates the PDF bytes for an invoice — same layout as GenerateInvoicePDF.
// This is extracted so both the PDF download and the email handler can use it.
func buildInvoicePDF(invoice *domain.Invoice, company *domain.Company) ([]byte, error) {
	return renderInvoicePDF(invoice, company)
}
