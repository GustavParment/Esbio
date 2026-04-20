package handlers

import (
	"esbio/api/internal/service"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReceiptHandler struct {
	receiptService *service.ReceiptService
}

func NewReceiptHandler(receiptService *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{receiptService: receiptService}
}

// Scan handles POST /receipts/scan
func (h *ReceiptHandler) Scan(c *gin.Context) {
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no receipt file uploaded"})
		return
	}
	defer file.Close()

	// Validate file size (max 20MB to accommodate multi-page PDFs)
	if header.Size > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 20MB"})
		return
	}

	// Validate MIME type
	mimeType := header.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/webp":      true,
		"image/heic":      true,
		"image/heif":      true,
		"application/pdf": true,
	}
	if !validTypes[mimeType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type, must be JPEG, PNG, WebP, HEIC or PDF"})
		return
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	result, err := h.receiptService.ScanReceipt(imageData, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan receipt"})
		return
	}

	c.JSON(http.StatusOK, result)
}
