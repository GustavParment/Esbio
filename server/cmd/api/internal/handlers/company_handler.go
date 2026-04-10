package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	companyService *service.CompanyService
}

func NewCompanyHandler(companyService *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService}
}

// ListCompanies handles GET /companies
func (h *CompanyHandler) ListCompanies(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	companies, err := h.companyService.GetCompaniesByUserID(userID.(int))
	if err != nil {
		internalError(c, err)
		return
	}
	if companies == nil {
		companies = []*domain.Company{}
	}

	// Add trial_days_left for free plan companies
	type companyWithTrial struct {
		*domain.Company
		TrialDaysLeft *int `json:"trial_days_left,omitempty"`
	}
	result := make([]companyWithTrial, len(companies))
	for i, comp := range companies {
		result[i] = companyWithTrial{Company: comp}
		if comp.Plan == "free" {
			if createdAt, err := time.Parse(time.RFC3339, comp.CreatedAt); err == nil {
				daysLeft := 30 - int(time.Since(createdAt).Hours()/24)
				if daysLeft < 0 {
					daysLeft = 0
				}
				result[i].TrialDaysLeft = &daysLeft
			}
		}
	}
	c.JSON(http.StatusOK, result)
}

// CreateCompany handles POST /companies
func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req struct {
		CompanyName string `json:"company_name" binding:"required"`
		OrgNumber   string `json:"org_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_name is required"})
		return
	}

	// Check if org number is already registered
	if req.OrgNumber != "" {
		existing, _ := h.companyService.GetCompanyByOrgNumber(req.OrgNumber)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Ett företag med detta organisationsnummer finns redan"})
			return
		}
	}

	company := &domain.Company{
		CompanyName: req.CompanyName,
		OrgNumber:   req.OrgNumber,
		CreatedBy:   userID.(int),
		Plan:        "free",
	}

	if err := h.companyService.CreateCompany(company); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, company)
}

// SelectCompany handles POST /companies/select
func (h *CompanyHandler) SelectCompany(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req struct {
		CompanyID int `json:"company_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	// Verify ownership
	owns, err := h.companyService.UserOwnsCompany(userID.(int), req.CompanyID)
	if err != nil || !owns {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Set company_id cookie
	setCookie(c, "company_id", strconv.Itoa(req.CompanyID), 604800)

	company, _ := h.companyService.GetCompanyByID(req.CompanyID)
	c.JSON(http.StatusOK, gin.H{"message": "company selected", "company": company})
}

// UpdateCompany handles PUT /companies/:id
func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	companyID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company ID"})
		return
	}

	// Verify ownership
	owns, _ := h.companyService.UserOwnsCompany(userID.(int), companyID)
	if !owns {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		CompanyName string `json:"company_name"`
		OrgNumber   string `json:"org_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company, _ := h.companyService.GetCompanyByID(companyID)
	if req.CompanyName != "" {
		company.CompanyName = req.CompanyName
	}
	if req.OrgNumber != "" {
		company.OrgNumber = req.OrgNumber
	}

	if err := h.companyService.UpdateCompany(company); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, company)
}

// DeleteCompany handles DELETE /companies/:id
func (h *CompanyHandler) DeleteCompany(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	companyID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company ID"})
		return
	}

	owns, _ := h.companyService.UserOwnsCompany(userID.(int), companyID)
	if !owns {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.companyService.DeleteCompany(companyID); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "company deleted"})
}
