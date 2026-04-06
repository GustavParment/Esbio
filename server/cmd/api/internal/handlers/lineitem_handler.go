package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LineItemHandler struct {
	lineItemService *service.LineItemService
}

func NewLineItemHandler(lineItemService *service.LineItemService) *LineItemHandler {
	return &LineItemHandler{
		lineItemService: lineItemService,
	}
}

// CreateLineItem handles POST /lineitems
func (h *LineItemHandler) CreateLineItem(c *gin.Context) {
	var lineItem domain.LineItem

	if err := c.ShouldBindJSON(&lineItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.lineItemService.CreateLineItem(&lineItem); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, lineItem)
}

// GetLineItemByID handles GET /lineitems/:id
func (h *LineItemHandler) GetLineItemByID(c *gin.Context) {
	idParam := c.Param("id")
	lineID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line item ID"})
		return
	}

	lineItem, err := h.lineItemService.GetLineItemByID(lineID)
	if err != nil {
		notFoundError(c, err)
		return
	}

	c.JSON(http.StatusOK, lineItem)
}

// GetLineItemsByVoucherID handles GET /lineitems/voucher/:voucherId
func (h *LineItemHandler) GetLineItemsByVoucherID(c *gin.Context) {
	voucherIDParam := c.Param("voucherId")
	voucherID, err := strconv.Atoi(voucherIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	lineItems, err := h.lineItemService.GetLineItemsByVoucherID(voucherID)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, lineItems)
}

// GetLineItemsByAccountNo handles GET /lineitems/account/:accountNo
func (h *LineItemHandler) GetLineItemsByAccountNo(c *gin.Context) {
	accountNoParam := c.Param("accountNo")
	accountNo, err := strconv.Atoi(accountNoParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number"})
		return
	}

	lineItems, err := h.lineItemService.GetLineItemsByAccountNo(accountNo)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, lineItems)
}

// UpdateLineItem handles PUT /lineitems/:id
func (h *LineItemHandler) UpdateLineItem(c *gin.Context) {
	idParam := c.Param("id")
	lineID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line item ID"})
		return
	}

	var lineItem domain.LineItem
	if err := c.ShouldBindJSON(&lineItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lineItem.LineID = lineID

	if err := h.lineItemService.UpdateLineItem(&lineItem); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "line item updated successfully",
		"lineItem": lineItem,
	})
}

// DeleteLineItem handles DELETE /lineitems/:id
func (h *LineItemHandler) DeleteLineItem(c *gin.Context) {
	idParam := c.Param("id")
	lineID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line item ID"})
		return
	}

	if err := h.lineItemService.DeleteLineItem(lineID); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "line item deleted successfully"})
}
