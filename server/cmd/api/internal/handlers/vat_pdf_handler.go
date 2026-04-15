package handlers

import (
	"bytes"
	"esbio/api/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/shopspring/decimal"
)

type VATPDFHandler struct {
	reportService  *service.ReportService
	companyService *service.CompanyService
}

func NewVATPDFHandler(reportService *service.ReportService, companyService *service.CompanyService) *VATPDFHandler {
	return &VATPDFHandler{
		reportService:  reportService,
		companyService: companyService,
	}
}

// GenerateVATDeclarationPDF handles GET /reports/vat/pdf?from_date=&to_date=
// Produces a Skatteverket SKV 4700-styled momsdeklaration the user can transcribe.
func (h *VATPDFHandler) GenerateVATDeclarationPDF(c *gin.Context) {
	cid, ok := getCompanyID(c)
	if !ok {
		return
	}

	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")
	if fromDate == "" || toDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_date and to_date query parameters are required"})
		return
	}

	report, err := h.reportService.GetVATReport(fromDate, toDate, cid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company, err := h.companyService.GetCompanyByID(cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load company"})
		return
	}

	// Map VAT entries to SKV 4700 boxes (small-company case: SE-domestic sales only)
	box05 := decimal.Zero // Försäljning som beskattas i Sverige (ex moms)
	box10 := decimal.Zero // Utgående moms 25%
	box11 := decimal.Zero // Utgående moms 12%
	box12 := decimal.Zero // Utgående moms 6%
	box42 := decimal.Zero // Övrig försäljning m.m. (0% / momsfri)

	for _, e := range report.Entries {
		switch e.TaxCode {
		case 25:
			box05 = box05.Add(e.TotalSales)
			box10 = box10.Add(e.TotalVAT)
		case 12:
			box05 = box05.Add(e.TotalSales)
			box11 = box11.Add(e.TotalVAT)
		case 6:
			box05 = box05.Add(e.TotalSales)
			box12 = box12.Add(e.TotalVAT)
		case 0:
			box42 = box42.Add(e.TotalSales)
		}
	}

	box48 := report.TotalInputVAT // Ingående moms att dra av
	box49 := report.NetVAT        // Moms att betala (positiv) / få tillbaka (negativ)

	// Build PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Mervärdesskattedeklaration (SKV 4700)"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.Cell(0, 5, tr("Underlag — överför värdena till Skatteverkets e-tjänst eller pappersblankett."))
	pdf.Ln(8)
	pdf.SetTextColor(0, 0, 0)

	// Company / period block
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, tr("Företag:"))
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, tr(company.CompanyName))
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, tr("Organisationsnummer:"))
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, tr(company.OrgNumber))
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, tr("Redovisningsperiod:"))
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("%s — %s", report.Period.FromDate, report.Period.ToDate))
	pdf.Ln(10)

	// Section A — Momspliktig försäljning
	drawSectionHeader(pdf, tr, "A. Momspliktig försäljning")
	drawBoxRow(pdf, tr, "05", "Försäljning som beskattas i Sverige", box05)
	drawBoxRow(pdf, tr, "42", "Övrig försäljning m.m.", box42)
	pdf.Ln(4)

	// Section B — Utgående moms
	drawSectionHeader(pdf, tr, "B. Utgående moms på försäljning i ruta 05")
	drawBoxRow(pdf, tr, "10", "Utgående moms 25%", box10)
	drawBoxRow(pdf, tr, "11", "Utgående moms 12%", box11)
	drawBoxRow(pdf, tr, "12", "Utgående moms 6%", box12)
	pdf.Ln(4)

	// Section H — Ingående moms
	drawSectionHeader(pdf, tr, "H. Ingående moms")
	drawBoxRow(pdf, tr, "48", "Ingående moms att dra av", box48)
	pdf.Ln(4)

	// Final result
	drawSectionHeader(pdf, tr, "Moms att betala eller få tillbaka")
	label := "Moms att betala (ruta 49)"
	if box49.IsNegative() {
		label = "Moms att få tillbaka (ruta 49)"
	}
	drawBoxRow(pdf, tr, "49", label, box49.Abs())

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 5, tr(fmt.Sprintf("Genererad av Esbio %s", time.Now().Format("2006-01-02 15:04"))))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	filename := fmt.Sprintf("momsdeklaration_%s_%s.pdf", fromDate, toDate)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

func drawSectionHeader(pdf *fpdf.Fpdf, tr func(string) string, title string) {
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(15, 29, 26)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(0, 8, " "+tr(title), "0", 0, "L", true, 0, "")
	pdf.Ln(8)
	pdf.SetTextColor(0, 0, 0)
}

func drawBoxRow(pdf *fpdf.Fpdf, tr func(string) string, boxNo, label string, value decimal.Decimal) {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(245, 245, 245)
	pdf.CellFormat(18, 8, tr("Ruta "+boxNo), "1", 0, "C", true, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(120, 8, " "+tr(label), "1", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 8, formatCurrency(value)+"  ", "1", 0, "R", false, 0, "")
	pdf.Ln(8)
}
