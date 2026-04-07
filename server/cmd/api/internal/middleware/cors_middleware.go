package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:3000":                                              true,
		"https://esbio-frontend-587000007432.europe-north1.run.app":         true,
		"https://esbio.se":                                                  true,
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// Allow exact matches and any local network origin (192.168.x.x:3000)
		isAllowed := allowedOrigins[origin] ||
			(strings.HasPrefix(origin, "http://192.168.") && strings.HasSuffix(origin, ":3000"))
		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
