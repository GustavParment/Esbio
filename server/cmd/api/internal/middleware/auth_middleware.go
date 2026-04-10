package middleware

import (
	"esbio/api/internal/auth"
	"esbio/api/internal/repository"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a middleware that validates JWT tokens from httpOnly cookies
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no authentication token found"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context for handlers to use
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole creates a middleware that checks if user has required role
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "user role not found"})
			c.Abort()
			return
		}

		userRole := role.(string)
		hasRole := false
		for _, r := range requiredRoles {
			if userRole == r {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserIDFromContext retrieves the user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

// GetUserRoleFromContext retrieves the user role from the Gin context
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	return role.(string), true
}

// CompanyMiddleware validates company_id cookie and verifies ownership
func CompanyMiddleware(companyRepo repository.CompanyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}

		companyIDStr, err := c.Cookie("company_id")
		if err != nil || companyIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected", "code": "NO_COMPANY"})
			c.Abort()
			return
		}

		companyID, err := strconv.Atoi(companyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
			c.Abort()
			return
		}

		owns, err := companyRepo.UserOwnsCompany(userID.(int), companyID)
		if err != nil || !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this company"})
			c.Abort()
			return
		}

		// Check trial expiration and plan status (skip for stripe routes — users must be able to pay)
		company, err := companyRepo.GetCompanyByID(companyID)
		isStripeRoute := strings.HasPrefix(c.Request.URL.Path, "/api/v1/stripe/")
		if err == nil && !isStripeRoute {
			// Block past_due subscriptions
			if company.PlanStatus == "past_due" {
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": "payment past due",
					"code":  "PAYMENT_PAST_DUE",
					"plan":  company.Plan,
				})
				c.Abort()
				return
			}

			// Check trial expiration for free plan
			if company.Plan == "free" {
				createdAt, parseErr := time.Parse(time.RFC3339, company.CreatedAt)
				if parseErr == nil {
					if time.Since(createdAt) > 30*24*time.Hour {
						c.JSON(http.StatusPaymentRequired, gin.H{
							"error": "trial expired",
							"code":  "TRIAL_EXPIRED",
							"plan":  company.Plan,
						})
						c.Abort()
						return
					}
				}
			}

			c.Set("companyName", company.CompanyName)
		}

		c.Set("companyID", companyID)
		c.Next()
	}
}

// GetCompanyIDFromContext retrieves the company ID from the Gin context
func GetCompanyIDFromContext(c *gin.Context) (int, bool) {
	companyID, exists := c.Get("companyID")
	if !exists {
		return 0, false
	}
	return companyID.(int), true
}
