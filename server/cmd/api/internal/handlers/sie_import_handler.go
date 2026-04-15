package handlers

import (
	"esbio/api/internal/service"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SIEImportHandler struct {
	importService *service.SIEImportService
}

func NewSIEImportHandler(importService *service.SIEImportService) *SIEImportHandler {
	return &SIEImportHandler{importService: importService}
}

const sieMaxUploadBytes = 25 * 1024 * 1024 // 25 MB

// PreviewSIE handles POST /reports/sie/import/preview
// Multipart form with field "file" — returns parse summary without writing.
func (h *SIEImportHandler) PreviewSIE(c *gin.Context) {
	if _, ok := getCompanyID(c); !ok {
		return
	}

	raw, err := readUploadedSIE(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	summary, err := h.importService.PreviewImport(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ImportSIE handles POST /reports/sie/import
// Multipart form with field "file" — parses and writes accounts + vouchers.
func (h *SIEImportHandler) ImportSIE(c *gin.Context) {
	cid, ok := getCompanyID(c)
	if !ok {
		return
	}
	uidVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	uid, ok := uidVal.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	raw, err := readUploadedSIE(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	summary, err := h.importService.Import(raw, cid, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func readUploadedSIE(c *gin.Context) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, sieMaxUploadBytes)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}
	if fileHeader.Size > sieMaxUploadBytes {
		return nil, http.ErrContentLength
	}
	f, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
