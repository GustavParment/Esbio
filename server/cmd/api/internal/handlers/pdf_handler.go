package handlers

import (
	"bytes"
	"esbio/api/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

type PDFHandler struct {
	voucherService *service.VoucherService
	accountService *service.AccountService
}

func NewPDFHandler(voucherService *service.VoucherService, accountService *service.AccountService) *PDFHandler {
	return &PDFHandler{
		voucherService: voucherService,
		accountService: accountService,
	}
}

// GenerateVoucherPDF handles GET /vouchers/:id/pdf
func (h *PDFHandler) GenerateVoucherPDF(c *gin.Context) {
	idParam := c.Param("id")
	voucherID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	// Get voucher with line items
	voucher, err := h.voucherService.GetVoucherByID(voucherID)
	if err != nil {
		notFoundError(c, err)
		return
	}

	// Create PDF with UTF-8 support
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up UTF-8 font translator
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Verifikat"))
	pdf.Ln(12)

	// Voucher number and date
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, tr(fmt.Sprintf("Verifikat #%d", voucher.VoucherNumber)))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(30, 6, tr("Datum:"))
	pdf.Cell(60, 6, voucher.Date.Time.Format("2006-01-02"))
	pdf.Ln(6)

	pdf.Cell(30, 6, tr("Period:"))
	pdf.Cell(60, 6, voucher.Period)
	pdf.Ln(6)

	if voucher.Reference != "" {
		pdf.Cell(30, 6, tr("Referens:"))
		pdf.Cell(60, 6, tr(voucher.Reference))
		pdf.Ln(6)
	}

	pdf.Cell(30, 6, tr("Beskrivning:"))
	pdf.MultiCell(150, 6, tr(voucher.Description), "", "", false)
	pdf.Ln(4)

	// Correction info
	if voucher.CorrectsVoucherID != nil {
		pdf.SetFont("Arial", "I", 9)
		pdf.SetTextColor(150, 0, 0)
		pdf.Cell(0, 6, tr(fmt.Sprintf("Detta verifikat rättar verifikat #%d", *voucher.CorrectsVoucherID)))
		pdf.Ln(6)
		pdf.SetTextColor(0, 0, 0)
	}
	if voucher.CorrectedByVoucherID != nil {
		pdf.SetFont("Arial", "I", 9)
		pdf.SetTextColor(150, 0, 0)
		pdf.Cell(0, 6, tr(fmt.Sprintf("Detta verifikat har rättats av verifikat #%d", *voucher.CorrectedByVoucherID)))
		pdf.Ln(6)
		pdf.SetTextColor(0, 0, 0)
	}

	pdf.Ln(4)

	// Line items table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(25, 8, tr("Konto"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(70, 8, tr("Kontonamn"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 8, tr("Debet"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 8, tr("Kredit"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(20, 8, tr("Moms"), "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	// Line items
	pdf.SetFont("Arial", "", 10)
	var totalDebit, totalCredit float64
	for _, line := range voucher.Lines {
		// Get account name
		accountName := ""
		if account, err := h.accountService.GetAccountByNo(line.AccountNo); err == nil {
			accountName = account.AccountName
		}

		pdf.CellFormat(25, 7, strconv.Itoa(line.AccountNo), "1", 0, "C", false, 0, "")
		pdf.CellFormat(70, 7, tr(truncateString(accountName, 35)), "1", 0, "L", false, 0, "")

		if line.DebitAmount > 0 {
			pdf.CellFormat(35, 7, formatCurrency(line.DebitAmount), "1", 0, "R", false, 0, "")
		} else {
			pdf.CellFormat(35, 7, "", "1", 0, "R", false, 0, "")
		}

		if line.CreditAmount > 0 {
			pdf.CellFormat(35, 7, formatCurrency(line.CreditAmount), "1", 0, "R", false, 0, "")
		} else {
			pdf.CellFormat(35, 7, "", "1", 0, "R", false, 0, "")
		}

		pdf.CellFormat(20, 7, fmt.Sprintf("%d%%", line.TaxCode), "1", 0, "C", false, 0, "")
		pdf.Ln(7)

		totalDebit += line.DebitAmount
		totalCredit += line.CreditAmount
	}

	// Totals row
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(95, 8, tr("Summa:"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 8, formatCurrency(totalDebit), "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 8, formatCurrency(totalCredit), "1", 0, "R", true, 0, "")
	pdf.CellFormat(20, 8, "", "1", 0, "C", true, 0, "")
	pdf.Ln(12)

	// Balance check
	if totalDebit == totalCredit {
		pdf.SetTextColor(0, 128, 0)
		pdf.Cell(0, 6, tr("Verifikatet är balanserat"))
	} else {
		pdf.SetTextColor(255, 0, 0)
		pdf.Cell(0, 6, tr("VARNING: Verifikatet är INTE balanserat!"))
	}
	pdf.SetTextColor(0, 0, 0)

	// Footer
	pdf.Ln(20)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 5, tr("Genererad av Esbio Bokföringssystem"))

	// Output PDF to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	// Set headers and send PDF
	filename := fmt.Sprintf("verifikat_%d.pdf", voucher.VoucherNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("%.2f kr", amount)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
