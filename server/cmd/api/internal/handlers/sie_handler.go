package handlers

import (
	"cmd/api/internal/service"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SIEHandler struct {
	sieGenerator *service.SIEGenerator
}

func NewSIEHandler(sieGenerator *service.SIEGenerator) *SIEHandler {
	return &SIEHandler{sieGenerator: sieGenerator}
}

// ExportSIE handles GET /reports/sie
func (h *SIEHandler) ExportSIE(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	if fromDate == "" || toDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_date and to_date query parameters are required"})
		return
	}

	sieBytes, err := h.sieGenerator.GenerateSIE4(fromDate, toDate, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("bokforing_%s_%s.se", strings.ReplaceAll(fromDate, "-", ""), strings.ReplaceAll(toDate, "-", ""))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", sieBytes)
}
