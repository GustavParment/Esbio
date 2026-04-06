package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Generic error messages for clients
const (
	errInternal   = "an internal error occurred"
	errNotFound   = "resource not found"
	errBadRequest = "invalid request"
)

// internalError logs the real error and returns a generic message to the client.
func internalError(c *gin.Context, err error) {
	log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": errInternal})
}

// notFoundError logs and returns a 404 with generic message.
func notFoundError(c *gin.Context, err error) {
	log.Printf("[WARN] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusNotFound, gin.H{"error": errNotFound})
}
