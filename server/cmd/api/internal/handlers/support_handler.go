package handlers

import (
	"esbio/api/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SupportHandler struct {
	emailService   *service.EmailService
	companyService *service.CompanyService
}

func NewSupportHandler(emailService *service.EmailService, companyService *service.CompanyService) *SupportHandler {
	return &SupportHandler{
		emailService:   emailService,
		companyService: companyService,
	}
}

// SendSupportMessage handles POST /support
func (h *SupportHandler) SendSupportMessage(c *gin.Context) {
	var req struct {
		Subject string `json:"subject" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ämne och meddelande krävs"})
		return
	}

	userID, _ := c.Get("userID")
	email, _ := c.Get("email")
	companyID, _ := c.Get("companyID")

	companyName := ""
	if companyID != nil {
		company, err := h.companyService.GetCompanyByID(companyID.(int))
		if err == nil {
			companyName = company.CompanyName
		}
	}

	emailStr, _ := email.(string)

	emailID, err := h.emailService.SendSupportEmail(
		emailStr,
		emailStr,
		companyName,
		req.Subject,
		req.Message,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to send support email from user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kunde inte skicka meddelandet"})
		return
	}

	log.Printf("[INFO] Support message from user %d (email_id: %s)", userID, emailID)

	c.JSON(http.StatusOK, gin.H{"message": "Tack! Vi återkommer så snart vi kan."})
}
