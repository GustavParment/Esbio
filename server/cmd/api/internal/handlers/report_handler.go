package handlers

import (
	"esbio/api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func getCompanyID(c *gin.Context) (int, bool) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return 0, false
	}
	return companyID.(int), true
}

// GetIncomeStatement handles GET /reports/income-statement
func (h *ReportHandler) GetIncomeStatement(c *gin.Context) {
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

	statement, err := h.reportService.GetIncomeStatement(fromDate, toDate, cid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statement)
}

// GetBalanceSheet handles GET /reports/balance-sheet
func (h *ReportHandler) GetBalanceSheet(c *gin.Context) {
	cid, ok := getCompanyID(c)
	if !ok {
		return
	}

	asOfDate := c.Query("as_of_date")

	if asOfDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "as_of_date query parameter is required"})
		return
	}

	sheet, err := h.reportService.GetBalanceSheet(asOfDate, cid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sheet)
}

// GetVATReport handles GET /reports/vat
func (h *ReportHandler) GetVATReport(c *gin.Context) {
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

	c.JSON(http.StatusOK, report)
}
