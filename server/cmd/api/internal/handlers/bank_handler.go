package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BankHandler struct {
	tinkService *service.TinkService
}

func NewBankHandler(tinkService *service.TinkService) *BankHandler {
	return &BankHandler{tinkService: tinkService}
}

// Connect initiates a Tink Link flow — returns the URL to redirect the user to.
func (h *BankHandler) Connect(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	userID, _ := c.Get("userID")

	url, err := h.tinkService.InitiateConnect(companyID.(int), userID.(int))
	if err != nil {
		log.Printf("Tink connect error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate bank connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// Callback processes the OAuth callback from Tink Link (called by the frontend after redirect).
func (h *BankHandler) Callback(c *gin.Context) {
	var req struct {
		State string `json:"state" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state and code are required"})
		return
	}

	companyID, err := h.tinkService.HandleCallback(req.State, req.Code)
	if err != nil {
		log.Printf("Tink callback error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process bank connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "company_id": companyID})
}

// ListConnections returns all bank connections for the current company.
func (h *BankHandler) ListConnections(c *gin.Context) {
	companyID, _ := c.Get("companyID")

	connections, err := h.tinkService.GetConnections(companyID.(int))
	if err != nil {
		log.Printf("Failed to list bank connections: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list connections"})
		return
	}

	if connections == nil {
		connections = []*domain.BankConnection{}
	}

	c.JSON(http.StatusOK, connections)
}

// ListAccounts returns all bank accounts for the current company.
func (h *BankHandler) ListAccounts(c *gin.Context) {
	companyID, _ := c.Get("companyID")

	accounts, err := h.tinkService.GetAccounts(companyID.(int))
	if err != nil {
		log.Printf("Failed to list bank accounts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// ListTransactions returns bank transactions with optional status filter and pagination.
func (h *BankHandler) ListTransactions(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transactions, err := h.tinkService.GetTransactions(companyID.(int), status, limit, offset)
	if err != nil {
		log.Printf("Failed to list bank transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// SyncTransactions triggers a manual sync of transactions from Tink.
func (h *BankHandler) SyncTransactions(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	connectionID, err := strconv.Atoi(c.Param("connectionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	newCount, err := h.tinkService.SyncTransactions(connectionID, companyID.(int))
	if err != nil {
		log.Printf("Tink sync error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"new_count": newCount})
}

// MapAccount maps a bank account to a BAS chart of accounts number.
func (h *BankHandler) MapAccount(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	bankAccountID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bank account ID"})
		return
	}

	var req struct {
		AccountNo int `json:"account_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_no is required"})
		return
	}

	if err := h.tinkService.MapAccount(bankAccountID, companyID.(int), req.AccountNo); err != nil {
		log.Printf("Failed to map bank account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to map account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ImportTransaction creates a voucher from a bank transaction.
func (h *BankHandler) ImportTransaction(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	txnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var req struct {
		AccountNo   int    `json:"account_no" binding:"required"`
		Description string `json:"description" binding:"required"`
		TaxCode     int    `json:"tax_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_no and description are required"})
		return
	}

	voucherID, err := h.tinkService.ImportTransaction(txnID, companyID.(int), req.AccountNo, req.Description, req.TaxCode)
	if err != nil {
		log.Printf("Failed to import bank transaction: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"voucher_id": voucherID})
}

// SkipTransaction marks a bank transaction as skipped.
func (h *BankHandler) SkipTransaction(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	txnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	if err := h.tinkService.SkipTransaction(txnID, companyID.(int)); err != nil {
		log.Printf("Failed to skip bank transaction: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DisconnectBank removes a bank connection.
func (h *BankHandler) DisconnectBank(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	connectionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	if err := h.tinkService.DisconnectBank(connectionID, companyID.(int)); err != nil {
		log.Printf("Failed to disconnect bank: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
