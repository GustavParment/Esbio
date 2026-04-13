package handlers

import (
	"esbio/api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InvoiceSettingsHandler struct {
	settingsService *service.InvoiceSettingsService
}

func NewInvoiceSettingsHandler(s *service.InvoiceSettingsService) *InvoiceSettingsHandler {
	return &InvoiceSettingsHandler{settingsService: s}
}

// GetSettings handles GET /invoices/settings
func (h *InvoiceSettingsHandler) GetSettings(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	settings, err := h.settingsService.GetByCompanyID(companyID.(int))
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings handles PUT /invoices/settings
func (h *InvoiceSettingsHandler) UpdateSettings(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	existing, err := h.settingsService.GetByCompanyID(companyID.(int))
	if err != nil {
		internalError(c, err)
		return
	}

	var req struct {
		Bankgiro                string `json:"bankgiro"`
		Plusgiro                 string `json:"plusgiro"`
		Swish                   string `json:"swish"`
		IBAN                    string `json:"iban"`
		BIC                     string `json:"bic"`
		FSkattText              string `json:"f_skatt_text"`
		DefaultPaymentTermsDays int    `json:"default_payment_terms_days"`
		InvoicePrefix           string `json:"invoice_prefix"`
		DefaultRevenueAccount   int    `json:"default_revenue_account"`
		DefaultPaymentAccount   int    `json:"default_payment_account"`
		FooterText              string `json:"footer_text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.Bankgiro = req.Bankgiro
	existing.Plusgiro = req.Plusgiro
	existing.Swish = req.Swish
	existing.IBAN = req.IBAN
	existing.BIC = req.BIC
	if req.FSkattText != "" {
		existing.FSkattText = req.FSkattText
	}
	if req.DefaultPaymentTermsDays > 0 {
		existing.DefaultPaymentTermsDays = req.DefaultPaymentTermsDays
	}
	existing.InvoicePrefix = req.InvoicePrefix
	if req.DefaultRevenueAccount > 0 {
		existing.DefaultRevenueAccount = req.DefaultRevenueAccount
	}
	if req.DefaultPaymentAccount > 0 {
		existing.DefaultPaymentAccount = req.DefaultPaymentAccount
	}
	existing.FooterText = req.FooterText

	if err := h.settingsService.Update(existing); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, existing)
}
