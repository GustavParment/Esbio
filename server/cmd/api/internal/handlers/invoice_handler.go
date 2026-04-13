package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// CreateInvoice handles POST /invoices
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	userID, _ := c.Get("userID")

	var req service.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.invoiceService.CreateInvoice(companyID.(int), userID.(int), req)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

// GetInvoiceByID handles GET /invoices/:id
func (h *InvoiceHandler) GetInvoiceByID(c *gin.Context) {
	invoiceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
		return
	}

	invoice, err := h.invoiceService.GetInvoiceWithDetails(invoiceID)
	if err != nil {
		notFoundError(c, err)
		return
	}

	companyID, _ := c.Get("companyID")
	if invoice.CompanyID != companyID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// ListInvoices handles GET /invoices
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	status := c.Query("status")

	var invoices []*domain.Invoice
	var err error

	if status != "" {
		invoices, err = h.invoiceService.GetInvoicesByStatus(companyID.(int), status)
	} else {
		invoices, err = h.invoiceService.GetInvoicesByCompanyID(companyID.(int))
	}

	if err != nil {
		internalError(c, err)
		return
	}
	if invoices == nil {
		invoices = []*domain.Invoice{}
	}
	c.JSON(http.StatusOK, invoices)
}

// FinalizeInvoice handles POST /invoices/:id/finalize
func (h *InvoiceHandler) FinalizeInvoice(c *gin.Context) {
	invoiceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
		return
	}
	userID, _ := c.Get("userID")

	invoice, err := h.invoiceService.FinalizeInvoice(invoiceID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// MarkAsPaid handles POST /invoices/:id/pay
func (h *InvoiceHandler) MarkAsPaid(c *gin.Context) {
	invoiceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
		return
	}
	userID, _ := c.Get("userID")

	var req struct {
		PaymentDate    string `json:"payment_date" binding:"required"`
		PaymentAccount int    `json:"payment_account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_date is required"})
		return
	}

	invoice, err := h.invoiceService.MarkAsPaid(invoiceID, userID.(int), req.PaymentDate, req.PaymentAccount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// CancelInvoice handles POST /invoices/:id/cancel
func (h *InvoiceHandler) CancelInvoice(c *gin.Context) {
	invoiceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
		return
	}
	userID, _ := c.Get("userID")

	invoice, err := h.invoiceService.CancelInvoice(invoiceID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// DeleteInvoice handles DELETE /invoices/:id
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	invoiceID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
		return
	}

	if err := h.invoiceService.DeleteInvoice(invoiceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice deleted"})
}
