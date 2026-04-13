package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

// CreateCustomer handles POST /customers
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	var req struct {
		Name                  string `json:"name" binding:"required"`
		OrgNumber             string `json:"org_number"`
		VATNumber             string `json:"vat_number"`
		AddressLine1          string `json:"address_line1"`
		AddressLine2          string `json:"address_line2"`
		PostalCode            string `json:"postal_code"`
		City                  string `json:"city"`
		Country               string `json:"country"`
		Email                 string `json:"email"`
		Phone                 string `json:"phone"`
		PaymentTermsDays      int    `json:"payment_terms_days"`
		DefaultRevenueAccount *int   `json:"default_revenue_account"`
		Notes                 string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	customer := &domain.Customer{
		CompanyID:             companyID.(int),
		Name:                  req.Name,
		OrgNumber:             req.OrgNumber,
		VATNumber:             req.VATNumber,
		AddressLine1:          req.AddressLine1,
		AddressLine2:          req.AddressLine2,
		PostalCode:            req.PostalCode,
		City:                  req.City,
		Country:               req.Country,
		Email:                 req.Email,
		Phone:                 req.Phone,
		PaymentTermsDays:      req.PaymentTermsDays,
		DefaultRevenueAccount: req.DefaultRevenueAccount,
		Notes:                 req.Notes,
	}

	if err := h.customerService.CreateCustomer(customer); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, customer)
}

// GetCustomerByID handles GET /customers/:id
func (h *CustomerHandler) GetCustomerByID(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	customer, err := h.customerService.GetCustomerByID(customerID)
	if err != nil {
		notFoundError(c, err)
		return
	}

	c.JSON(http.StatusOK, customer)
}

// ListCustomers handles GET /customers
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	customers, err := h.customerService.GetCustomersByCompanyID(companyID.(int))
	if err != nil {
		internalError(c, err)
		return
	}

	if customers == nil {
		customers = []*domain.Customer{}
	}
	c.JSON(http.StatusOK, customers)
}

// SearchCustomers handles GET /customers/search?q=
func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	query := c.Query("q")
	customers, err := h.customerService.SearchCustomers(companyID.(int), query)
	if err != nil {
		internalError(c, err)
		return
	}

	if customers == nil {
		customers = []*domain.Customer{}
	}
	c.JSON(http.StatusOK, customers)
}

// UpdateCustomer handles PUT /customers/:id
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	customerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	existing, err := h.customerService.GetCustomerByID(customerID)
	if err != nil {
		notFoundError(c, err)
		return
	}
	if existing.CompanyID != companyID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		Name                  string `json:"name"`
		OrgNumber             string `json:"org_number"`
		VATNumber             string `json:"vat_number"`
		AddressLine1          string `json:"address_line1"`
		AddressLine2          string `json:"address_line2"`
		PostalCode            string `json:"postal_code"`
		City                  string `json:"city"`
		Country               string `json:"country"`
		Email                 string `json:"email"`
		Phone                 string `json:"phone"`
		PaymentTermsDays      int    `json:"payment_terms_days"`
		DefaultRevenueAccount *int   `json:"default_revenue_account"`
		Notes                 string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.OrgNumber = req.OrgNumber
	existing.VATNumber = req.VATNumber
	existing.AddressLine1 = req.AddressLine1
	existing.AddressLine2 = req.AddressLine2
	existing.PostalCode = req.PostalCode
	existing.City = req.City
	if req.Country != "" {
		existing.Country = req.Country
	}
	existing.Email = req.Email
	existing.Phone = req.Phone
	if req.PaymentTermsDays > 0 {
		existing.PaymentTermsDays = req.PaymentTermsDays
	}
	existing.DefaultRevenueAccount = req.DefaultRevenueAccount
	existing.Notes = req.Notes

	if err := h.customerService.UpdateCustomer(existing); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteCustomer handles DELETE /customers/:id
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	customerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	existing, err := h.customerService.GetCustomerByID(customerID)
	if err != nil {
		notFoundError(c, err)
		return
	}
	if existing.CompanyID != companyID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.customerService.DeleteCustomer(customerID); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}
