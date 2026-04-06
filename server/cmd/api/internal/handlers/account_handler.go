package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// CreateAccount handles POST /accounts
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var account domain.Account
	
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.accountService.CreateAccount(&account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "account created successfully",
		"account": account,
	})
}

// GetAccountByNo handles GET /accounts/:accountNo
func (h *AccountHandler) GetAccountByNo(c *gin.Context) {
	accountNoParam := c.Param("accountNo")
	accountNo, err := strconv.Atoi(accountNoParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number"})
		return
	}

	account, err := h.accountService.GetAccountByNo(accountNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetAllAccounts handles GET /accounts
func (h *AccountHandler) GetAllAccounts(c *gin.Context) {
	accounts, err := h.accountService.GetAllAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetAccountsByGroup handles GET /accounts/group/:group
func (h *AccountHandler) GetAccountsByGroup(c *gin.Context) {
	groupParam := c.Param("group")
	accountGroup, err := strconv.Atoi(groupParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account group"})
		return
	}

	accounts, err := h.accountService.GetAccountsByGroup(accountGroup)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
		"group":    accountGroup,
		"count":    len(accounts),
	})
}

// UpdateAccount handles PUT /accounts/:accountNo
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	accountNoParam := c.Param("accountNo")
	accountNo, err := strconv.Atoi(accountNoParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number"})
		return
	}

	var account domain.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.AccountNo = accountNo

	if err := h.accountService.UpdateAccount(&account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account updated successfully",
		"account": account,
	})
}

// DeleteAccount handles DELETE /accounts/:accountNo
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	accountNoParam := c.Param("accountNo")
	accountNo, err := strconv.Atoi(accountNoParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number"})
		return
	}

	if err := h.accountService.DeleteAccount(accountNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

// GetAccountLedger handles GET /accounts/:accountNo/ledger
func (h *AccountHandler) GetAccountLedger(c *gin.Context) {
	accountNoParam := c.Param("accountNo")
	accountNo, err := strconv.Atoi(accountNoParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number"})
		return
	}

	// Get period from query parameter (optional)
	period := c.Query("period")

	entries, err := h.accountService.GetLedger(accountNo, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}
