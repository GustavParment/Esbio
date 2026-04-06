package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	voucherService *service.VoucherService
}

func NewVoucherHandler(voucherService *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{
		voucherService: voucherService,
	}
}

// CreateVoucher handles POST /vouchers
func (h *VoucherHandler) CreateVoucher(c *gin.Context) {
	var voucher domain.Voucher

	if err := c.ShouldBindJSON(&voucher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}
	voucher.CompanyID = companyID.(int)

	userID, exists := c.Get("userID")
	if exists {
		voucher.CreatedBy = userID.(int)
	}

	if err := h.voucherService.CreateVoucher(&voucher); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, voucher)
}

// GetVoucherByID handles GET /vouchers/:id
func (h *VoucherHandler) GetVoucherByID(c *gin.Context) {
	idParam := c.Param("id")
	voucherID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	voucher, err := h.voucherService.GetVoucherByID(voucherID)
	if err != nil {
		notFoundError(c, err)
		return
	}

	c.JSON(http.StatusOK, voucher)
}

// GetAllVouchers handles GET /vouchers
func (h *VoucherHandler) GetAllVouchers(c *gin.Context) {
	vouchers, err := h.voucherService.GetAllVouchers()
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, vouchers)
}

// GetVouchersByPeriod handles GET /vouchers/period/:period
func (h *VoucherHandler) GetVouchersByPeriod(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	period := c.Param("period")

	vouchers, err := h.voucherService.GetVouchersByPeriod(companyID.(int), period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vouchers)
}

// GetVouchersByCompanyID handles GET /vouchers/company
func (h *VoucherHandler) GetVouchersByCompanyID(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	vouchers, err := h.voucherService.GetVouchersByCompanyID(companyID.(int))
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, vouchers)
}

// GetAllPeriods handles GET /vouchers/periods
func (h *VoucherHandler) GetAllPeriods(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	periods, err := h.voucherService.GetAllPeriods(companyID.(int))
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, periods)
}

// UpdateVoucher handles PUT /vouchers/:id
func (h *VoucherHandler) UpdateVoucher(c *gin.Context) {
	idParam := c.Param("id")
	voucherID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	var voucher domain.Voucher
	if err := c.ShouldBindJSON(&voucher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	voucher.VoucherID = voucherID

	if err := h.voucherService.UpdateVoucher(&voucher); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "voucher updated successfully",
		"voucher": voucher,
	})
}

// DeleteVoucher handles DELETE /vouchers/:id
func (h *VoucherHandler) DeleteVoucher(c *gin.Context) {
	idParam := c.Param("id")
	voucherID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	if err := h.voucherService.DeleteVoucher(voucherID); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "voucher and associated line items deleted successfully"})
}

// ValidateVoucherBalance handles GET /vouchers/:id/validate
func (h *VoucherHandler) ValidateVoucherBalance(c *gin.Context) {
	idParam := c.Param("id")
	voucherID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	balanced, err := h.voucherService.ValidateVoucherBalance(voucherID)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"voucherId": voucherID,
		"balanced":  balanced,
		"message":   map[bool]string{true: "Voucher is balanced", false: "Voucher is NOT balanced"}[balanced],
	})
}

// CreateCorrectionVoucher handles POST /vouchers/:id/correct
func (h *VoucherHandler) CreateCorrectionVoucher(c *gin.Context) {
	idParam := c.Param("id")
	voucherID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	// Get user ID from request body
	var req struct {
		UserID int `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	correctionVoucher, err := h.voucherService.CreateCorrectionVoucher(voucherID, req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, correctionVoucher)
}

// CreateCorrectionWithChanges handles POST /vouchers/:id/correct-with-changes
func (h *VoucherHandler) CreateCorrectionWithChanges(c *gin.Context) {
	idParam := c.Param("id")
	voucherID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	// Get request body with new voucher data
	var req struct {
		UserID int `json:"user_id" binding:"required"`
		NewVoucher struct {
			Date        string  `json:"date"`
			Description string  `json:"description"`
			Reference   string  `json:"reference"`
			TotalAmount float64 `json:"total_amount"`
			Period      string  `json:"period"`
			CreatedBy   int     `json:"created_by"`
		} `json:"new_voucher" binding:"required"`
		NewLineItems []struct {
			AccountNo    int     `json:"account_no"`
			DebitAmount  float64 `json:"debit_amount"`
			CreditAmount float64 `json:"credit_amount"`
			TaxCode      int     `json:"tax_code"`
		} `json:"new_line_items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert line items to domain model
	newLineItems := make([]domain.LineItem, len(req.NewLineItems))
	for i, item := range req.NewLineItems {
		newLineItems[i] = domain.LineItem{
			AccountNo:    item.AccountNo,
			DebitAmount:  item.DebitAmount,
			CreditAmount: item.CreditAmount,
			TaxCode:      item.TaxCode,
		}
	}

	correctionVoucher, err := h.voucherService.CreateCorrectionWithChanges(
		voucherID,
		req.UserID,
		req.NewVoucher.Date,
		req.NewVoucher.Description,
		req.NewVoucher.Reference,
		req.NewVoucher.Period,
		newLineItems,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, correctionVoucher)
}
