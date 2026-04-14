package handlers

import (
	"bytes"
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/shopspring/decimal"
)

type InvoicePDFHandler struct {
	invoiceService *service.InvoiceService
	companyService *service.CompanyService
}

func NewInvoicePDFHandler(invoiceService *service.InvoiceService, companyService *service.CompanyService) *InvoicePDFHandler {
	return &InvoicePDFHandler{
		invoiceService: invoiceService,
		companyService: companyService,
	}
}

// GenerateInvoicePDF handles GET /invoices/:id/pdf
func (h *InvoicePDFHandler) GenerateInvoicePDF(c *gin.Context) {
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

	company, err := h.companyService.GetCompanyByID(invoice.CompanyID)
	if err != nil {
		internalError(c, err)
		return
	}

	pdfBytes, err := renderInvoicePDF(invoice, company)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	filename := fmt.Sprintf("faktura_%d.pdf", invoice.InvoiceNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// renderInvoicePDF generates PDF bytes for an invoice. Used by both the PDF download
// endpoint and the email handler.
func renderInvoicePDF(invoice *domain.Invoice, company *domain.Company) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// === HEADER ===
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(100, 12, tr(company.CompanyName))
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 12, "FAKTURA", "", 0, "R", false, 0, "")
	pdf.Ln(14)

	// Company info line
	pdf.SetFont("Arial", "", 9)
	companyInfo := ""
	if company.OrgNumber != "" {
		companyInfo += "Org.nr: " + company.OrgNumber
	}
	pdf.Cell(100, 5, tr(companyInfo))
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Fakturanr: %d", invoice.InvoiceNumber), "", 0, "R", false, 0, "")
	pdf.Ln(5)
	pdf.Cell(100, 5, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Fakturadatum: %s", invoice.InvoiceDate.Time.Format("2006-01-02")), "", 0, "R", false, 0, "")
	pdf.Ln(5)
	pdf.Cell(100, 5, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Forfallodatum: %s", invoice.DueDate.Time.Format("2006-01-02")), "", 0, "R", false, 0, "")
	pdf.Ln(5)
	pdf.Cell(100, 5, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("OCR: %s", invoice.OCRNumber), "", 0, "R", false, 0, "")
	pdf.Ln(12)

	// === CUSTOMER ===
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 6, tr("Kund:"))
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	if invoice.Customer != nil {
		cust := invoice.Customer
		pdf.Cell(0, 5, tr(cust.Name))
		pdf.Ln(5)
		if cust.AddressLine1 != "" {
			pdf.Cell(0, 5, tr(cust.AddressLine1))
			pdf.Ln(5)
		}
		if cust.AddressLine2 != "" {
			pdf.Cell(0, 5, tr(cust.AddressLine2))
			pdf.Ln(5)
		}
		if cust.PostalCode != "" || cust.City != "" {
			pdf.Cell(0, 5, tr(cust.PostalCode+" "+cust.City))
			pdf.Ln(5)
		}
		if cust.OrgNumber != "" {
			pdf.Cell(0, 5, tr("Org.nr: "+cust.OrgNumber))
			pdf.Ln(5)
		}
	}

	// References
	if invoice.YourReference != "" || invoice.OurReference != "" {
		pdf.Ln(3)
		pdf.SetFont("Arial", "", 9)
		if invoice.YourReference != "" {
			pdf.Cell(95, 5, tr("Er referens: "+invoice.YourReference))
		}
		if invoice.OurReference != "" {
			pdf.CellFormat(0, 5, tr("Var referens: "+invoice.OurReference), "", 0, "R", false, 0, "")
		}
		pdf.Ln(5)
	}

	pdf.Ln(8)

	// === LINE ITEMS TABLE ===
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(75, 7, tr("Beskrivning"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(18, 7, tr("Antal"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(15, 7, tr("Enhet"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 7, tr("A-pris"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(18, 7, tr("Moms"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, tr("Belopp"), "1", 0, "R", true, 0, "")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 9)
	for _, line := range invoice.Lines {
		desc := line.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		pdf.CellFormat(75, 6, tr(desc), "1", 0, "L", false, 0, "")
		pdf.CellFormat(18, 6, line.Quantity.String(), "1", 0, "R", false, 0, "")
		pdf.CellFormat(15, 6, tr(line.Unit), "1", 0, "C", false, 0, "")
		pdf.CellFormat(28, 6, fmtMoney(line.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(18, 6, fmt.Sprintf("%d%%", line.VATRate), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, fmtMoney(line.LineTotal), "1", 0, "R", false, 0, "")
		pdf.Ln(6)
	}

	pdf.Ln(4)

	// === TOTALS ===
	x := 120.0
	pdf.SetFont("Arial", "", 10)

	pdf.SetX(x)
	pdf.CellFormat(34, 6, tr("Netto:"), "", 0, "L", false, 0, "")
	pdf.CellFormat(30, 6, fmtMoney(invoice.Subtotal)+" kr", "", 0, "R", false, 0, "")
	pdf.Ln(6)

	// VAT summary per rate
	vatByRate := make(map[int]decimal.Decimal)
	for _, line := range invoice.Lines {
		vatByRate[line.VATRate] = vatByRate[line.VATRate].Add(line.VATAmount)
	}
	for rate, amount := range vatByRate {
		if rate == 0 {
			continue
		}
		pdf.SetX(x)
		pdf.CellFormat(34, 6, fmt.Sprintf("Moms %d%%:", rate), "", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, fmtMoney(amount)+" kr", "", 0, "R", false, 0, "")
		pdf.Ln(6)
	}

	if !invoice.Rounding.IsZero() {
		pdf.SetX(x)
		pdf.CellFormat(34, 6, tr("Oresavrundning:"), "", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, fmtMoney(invoice.Rounding)+" kr", "", 0, "R", false, 0, "")
		pdf.Ln(6)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.SetX(x)
	pdf.CellFormat(34, 8, tr("Att betala:"), "", 0, "L", false, 0, "")
	pdf.CellFormat(30, 8, fmtMoney(invoice.Total)+" kr", "", 0, "R", false, 0, "")
	pdf.Ln(12)

	// === PAYMENT INFO ===
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 6, tr("Betalningsinformation"))
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)

	if invoice.Bankgiro != "" {
		pdf.Cell(30, 5, "Bankgiro:")
		pdf.Cell(0, 5, invoice.Bankgiro)
		pdf.Ln(5)
	}
	if invoice.Plusgiro != "" {
		pdf.Cell(30, 5, "Plusgiro:")
		pdf.Cell(0, 5, invoice.Plusgiro)
		pdf.Ln(5)
	}
	pdf.Cell(30, 5, "OCR:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 5, invoice.OCRNumber)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(5)
	pdf.Cell(30, 5, tr("Forfallodatum:"))
	pdf.Cell(0, 5, invoice.DueDate.Time.Format("2006-01-02"))
	pdf.Ln(10)

	// === FOOTER ===
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	if invoice.FSkattText != "" {
		pdf.Cell(0, 5, tr(invoice.FSkattText))
		pdf.Ln(5)
	}
	pdf.Cell(0, 5, tr("Genererad av Esbio Bokforingssystem"))

	// Output
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func fmtMoney(d decimal.Decimal) string {
	return d.StringFixed(2)
}
